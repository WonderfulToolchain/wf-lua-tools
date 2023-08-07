-- SPDX-License-Identifier: MIT
-- SPDX-FileContributor: Adrian "asie" Siekierka, 2023

--- .c/.h file generator.
-- @module wf.internal.bin2c
-- @alias M

local M = {}

function M.bin2c(c_file, h_file, program_name, entries)
    local current_date = os.date()
    local comment_header = "// autogenerated by " .. program_name .. " on " .. current_date .. "\n\n"
    local c_values_per_line <const> = 12

    h_file:write(comment_header)
    h_file:write("#pragma once\n#include <stdint.h>\n#include <wonderful.h>\n\n")

    c_file:write(comment_header)
    c_file:write("#include <stdint.h>\n#include <wonderful.h>\n\n")

    for array_name, entry in pairs(entries) do
        local data = entry.data
        
        h_file:write("#define " .. array_name .. "_size (" .. #data .. ")\n")
        h_file:write("extern const uint8_t ")
        if entry.address_space then
            h_file:write(entry.address_space .. " ")
        end
        h_file:write(array_name .. "[" .. #data .. "];\n")
    
        c_file:write("const uint8_t ")
        if entry.address_space then
            c_file:write(entry.address_space .. " ")
        end
        c_file:write(array_name .. "[" .. #data .. "] ")
        if entry.align then
            c_file:write("__attribute__((aligned(" .. entry.align .. "))) ")
        end
        c_file:write("= {");
        for i = 1, #data do
            i_line = i % c_values_per_line
            if i_line == 1 then
                c_file:write("\n\t")
            else
                c_file:write(" ")
            end
            c_file:write(string.format("0x%02X", data:byte(i)))
            if i < #data then
                c_file:write(",")
            end
        end
        c_file:write("\n};\n");
    end
end

return M
