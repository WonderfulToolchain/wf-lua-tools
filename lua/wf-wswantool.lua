#!/opt/wonderful/bin/wf-lua
-- SPDX-License-Identifier: MIT
-- SPDX-FileContributor: Adrian "asie" Siekierka, 2023

local stringx = require('pl.stringx')
local tablex = require('pl.tablex')

--- Convert an user-provided memory layout to a linklayout.
-- For now at least, a linklayout is organized per-segment, with two segments: 0x0000 ("iram") and 0x1000 ("sram").
-- Each segment is a table with entries { key, start, end [inclusive] }.
-- In addition, the following fields are provided:
-- - ds: data segment value
-- - ss: stack segment value
-- - sp: stack pointer value
-- ROM areas are currently not handled by the linklayout.
function rom_memory_to_linklayout(memory)
    local layout = memory.layout
    local result = {["ds"]=nil, ["ss"]=nil, ["sp"]=nil, ["iram"]={}, ["sram"]={}}
    result.model = memory.model or "medium"

    -- validate layout keys
    for k, v in pairs(layout) do
        if k:sub(1,1):find("[^a-zA-Z]") or k:find("[^a-zA-Z0-9_]") then
            error("memory layout contains invalid identifier: " .. k)
        end
        if v[2] < v[1] then
            error("memory layout contains negative-sized entry: " .. k)
        end
        if v[1] >= 0x20000 then
            error("memory layout contains entry in ROM area: " .. k)
        end
        if (v[1] & 0xF0000) ~= (v[2] & 0xF0000) then
            error("memory layout contains entry across segment boundaries: " .. k)
        end

        if (k ~= "stack") then
            for k2, v2 in pairs(layout) do
                if (k ~= k2) and (k2 ~= "stack") and (v[1] <= v2[2]) and (v2[1] <= v[2]) then
                    error("memory layout contains overlapping entries: " .. k .. ", " .. k2)
                end
            end
        end
    end

    -- calculate DS
    if layout.c_heap then
        result.ds = (layout.c_heap[1] & 0xF0000) >> 4
    else
        result.ds = 0x0000
    end

    -- calculate SS/SP
    local stack_pointer = nil
    if layout.stack then
        stack_pointer = layout.stack[2] + 1
    elseif layout.c_heap then
        stack_pointer = layout.c_heap[2] + 1
    else
        error("missing stack area in memory layout")
    end
    result.ss = (stack_pointer & 0xF0000) >> 4
    result.sp = (stack_pointer & 0x0FFFF)

    -- arrange layout
    for k, v in tablex.sortv(layout, function(x, y)
        return x[1] < y[1]
    end) do
        if k ~= "stack" then
            local entry = {k, v[1], v[2]}
            if v[1] >= 0x10000 then
                table.insert(result.sram, entry)
            else
                table.insert(result.iram, entry)
            end
        end
    end

    return result
end

-- Abandon all hope, all ye who enter here.
function rom_write_linkscript(f, linklayout, constants, rom_start, rom_length)
    constants = tablex.copy(constants)
    constants["__wf_stack_pointer"] = linklayout.sp

    local is_far_text = linklayout.model == "medium" or linklayout.model == "large" or linklayout.model == "huge"
    local memory_regions = {
        ["IRAM"] = {0x00000, 0x10000, "wx"},
        ["SRAM"] = {0x10000, 0x10000, "wx"},
        ["ROM"] = {rom_start, rom_length, "rx"}
    }

    local function write_constant_symbol(key, value)
        if type(value) ~= "number" then
            error("invalid constant type for " .. key)
        end
        f:write("\n    \"" .. key .. "!\" = 0;\n")
        f:write("    \"" .. key .. "\" = ABSOLUTE(" .. value .. ");\n")
    end

    -- write header + memory regions
    f:write([[
/* automatically generated by wf-wswantool on ]] .. os.date() .. [[ */
OUTPUT_FORMAT("elf32-i386")
ENTRY(_start)
MEMORY
{
]])
    for k, v in pairs(memory_regions) do
        f:write("    " .. k:upper() .. " (" .. v[3] .. ") : ORIGIN = " .. v[1] .. ", LENGTH = " .. v[2] .. "\n")
    end
    f:write("}\nSECTIONS\n{\n")

    -- generate ROM region (.text)
    f:write([[
    /* ROM */
    
    ".text!" ]] .. rom_start .. [[ (NOLOAD) :
    {
        "__stext!" = .;
        KEEP(*(".start!"))
        *(".text!*" ".text.*!")
]])

    if not is_far_text then
        f:write([[
        *(".fartext!*" ".fartext.*!")
        *(".farrodata!*" ".farrodata.*!")
]])
    end

f:write([[
        "__etext!" = .;
    } >ROM

    .text . :
    {
        __stext = .;
        KEEP(*(".start"))
        *(.text ".text.*[^&]")
]])

    if not is_far_text then
        f:write([[
        *(".fartext.*[^&]")
        *(".farrodata.*[^&]")
]])
    end
    
    f:write([[
        __etext = .;
        . = ALIGN (16);
    }
]])

    -- add constant symbols here
    for k, v in pairs(constants) do
        write_constant_symbol(k, v)
    end

    f:write([[

    ".text&" . (NOLOAD) :
    {
        "__stext&" = .;
        KEEP(*(".start&"))
        *(".text&*" ".text.*&")
]])

    if not is_far_text then
        f:write([[
        *(".fartext&*" ".fartext.*&")
        *(".farrodata&*" ".farrodata.*&")
]])
    end

    f:write([[
        "__etext&" = .;
    }
]])
    if is_far_text then
        f:write([[

    .fartext ALIGN (0x10) : SUBALIGN (0x10) {
        *(SORT (".fartext!*"))
        *(SORT (".fartext$*"))
        *(SORT (".fartext&*"))
        *(SORT (".fartext.*"))
        . = .;
    }

    .farrodata ALIGN (0x10) : SUBALIGN (0x10) {
        *(SORT (".farrodata!*"))
        *(SORT (".farrodata$*"))
        *(SORT (".farrodata&*"))
        *(SORT (".farrodata.*"))
        . = .;
    }
]])
    end

    f:write([[

    .erom . (NOLOAD) :
    {
        . = ALIGN (16);
        "__erom" = .;
        "__erom!" = .;
        "__erom&" = .;
        . = .;
    }

]])
    local last_end_key = "erom"

    local function write_region(key, start, sections)
        local function write_region_segment_marker(symbol)
            local key_symbol = key .. symbol
            f:write("\n")
            if symbol == "!" then
                f:write("    \"." .. key_symbol .. "\" " .. start .. " (NOLOAD) : AT(ADDR(\"." .. last_end_key .. "\") + SIZEOF(\"." .. last_end_key .. "\"))\n")
            else
                f:write("    \"." .. key_symbol .. "\" . (NOLOAD) :\n")
            end
            f:write("    {\n")
            if symbol == "!" then
                for _, section in ipairs(sections) do
                    if section[1] == "c_heap" then
                        f:write("        \"__sdata!\" = .;\n")
                    else
                        f:write("        \"__s" .. section[1] .. "!\" = .;\n")
                    end
                end
            end
            for _, section in ipairs(sections) do
                if section[1] == "c_heap" then
                    f:write("        *(\".rodata" .. symbol .. "*\" \".rodata.*" .. symbol .. "\")\n")
                    f:write("        *(\".data" .. symbol .. "*\" \".data.*" .. symbol .. "\")\n")
                    f:write("        *(\".bss" .. symbol .. "*\" \".bss.*" .. symbol .. "\")\n")
                else
                    f:write("        *(\"." .. section[1] .. symbol .. "*\" \"." .. section[1] .. ".*" .. symbol .. "\")\n")
                end
            end
            if symbol == "!" then
                for _, section in ipairs(sections) do
                    if section[1] == "c_heap" then
                        f:write("        \"__edata!\" = .;\n")
                    else
                        f:write("        \"__e" .. section[1] .. "!\" = .;\n")
                    end
                end
            end
            f:write("    } >" .. key:upper() .. "\n")
        end
        
        f:write("    /* " .. key:upper() .. " */\n")
        if #sections > 0 then
            write_region_segment_marker("!")
            for _, section in ipairs(sections) do
                f:write("\n")
                f:write("    . = " .. section[2] .. ";\n")
                if section[1] == "c_heap" then
                    f:write([[
    __sheap = .;

    __sdata = .;
    ".data" . : AT(ADDR(".erom") + SIZEOF(".erom"))
    {
            *(.rodata ".rodata.*[^&]")
            *(.data ".data.*[^&]")
    }
    __edata = .;
    "__ldata!" = 0;
    __ldata = SIZEOF(.data);
    "__lwdata!" = 0;
    __lwdata = (__ldata + 1) / 2;

    __sbss = .;
    .bss . (NOLOAD) :
    {
            *(.bss ".bss.*[^&]")
    }
    __ebss = .;
    "__lbss!" = 0;
    __lbss = SIZEOF(.bss);
    "__lwbss!" = 0;
    __lwbss = (__lbss + 1) / 2;

    . = ]] .. (section[3] + 1) .. [[;
    "__eheap!" = 0;
    __eheap = .;
]])
                else
                    f:write([[
    __s]] .. section[1] .. [[ = .;
    .]] .. section[1] .. [[ . (NOLOAD) :
    {
            *(.]] .. section[1] .. [[ ".]] .. section[1] .. [[.*[^&]")
    }
    __e]] .. section[1] .. [[ = .;
    "__l]] .. section[1] .. [[!" = 0;
    __l]] .. section[1] .. [[ = SIZEOF(.]] .. section[1] .. [[);
    "__lw]] .. section[1] .. [[!" = 0;
    __lw]] .. section[1] .. [[ = (__l]] .. section[1] .. [[ + 1) / 2;
    . = ]] .. (section[3] + 1) .. [[;
    __e]] .. section[1] .. [[ = .;
]])
                end
            end
            write_region_segment_marker("&")

            last_end_key = "e" .. key
            f:write([[

    .]] .. last_end_key .. [[ . (NOLOAD) :
    {
        . = ALIGN (2);
        "__]] .. last_end_key .. [[" = .;
        "__]] .. last_end_key .. [[!" = .;
        "__]] .. last_end_key .. [[&" = .;
        . = .;
    }
]])
        end
        f:write("\n")
    end

    -- generate RAM regions
    write_region("iram", 0x00000, linklayout.iram)
    write_region("sram", 0x10000, linklayout.sram)

    -- generate debug sections
    local function write_dummy_section(...)
        local arg = {...}
        f:write("    \"" .. arg[1] .. "!\" = 0;\n")
        f:write("    " .. arg[1] .. " 0 : { *(" .. stringx.join(" ", arg) .. ") }\n")
    end

    -- DWARF
    write_dummy_section(".debug")
    write_dummy_section(".line")
    write_dummy_section(".debug_srcinfo")
    write_dummy_section(".debug_sfnames")
    write_dummy_section(".debug_aranges")
    write_dummy_section(".debug_pubnames")
    write_dummy_section(".debug_info", ".gnu.linkonce.wi.*")
    write_dummy_section(".debug_abbrev")
    write_dummy_section(".debug_line", ".debug_line.*", ".debug_line_end")
    write_dummy_section(".debug_frame")
    write_dummy_section(".debug_str")
    write_dummy_section(".debug_loc")
    write_dummy_section(".debug_macinfo")
    write_dummy_section(".debug_weaknames")
    write_dummy_section(".debug_funcnames")
    write_dummy_section(".debug_typenames")
    write_dummy_section(".debug_varnames")
    write_dummy_section(".debug_pubtypes")
    write_dummy_section(".debug_ranges")
    write_dummy_section(".debug_addr")
    write_dummy_section(".debug_line_str")
    write_dummy_section(".debug_loclists")
    write_dummy_section(".debug_macro")
    write_dummy_section(".debug_names")
    write_dummy_section(".debug_rnglists")
    write_dummy_section(".debug_str_offsets")
    write_dummy_section(".debug_sup")

    f:write("}\n")
end

commands = {}
commands.romlink = require('wf.internal.wswantool.romlink')

require('wf.internal.tool')