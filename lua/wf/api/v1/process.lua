-- SPDX-License-Identifier: MIT
-- SPDX-FileContributor: Adrian "asie" Siekierka, 2023

--- wf-process helper functions
-- @module wf.api.v1.process
-- @alias M

if _WFPROCESS == nil then
    error("not running inside wf-process")
end

local path = require("pl.path")
local wfutil = require("wf.internal.util")

local M = {}

local tmpfile_counter = 0
--- Allocate a temporary file.
-- This file will be deleted once wf-process finishes operation.
-- @tparam ?string ext File extension.
-- @treturn table Temporary file table.
function M.tmpfile(ext)
    tmpfile_counter = tmpfile_counter + 1
    return {["file"]=_WFPROCESS.temp_dir:path(string.format("wf%05d%s", tmpfile_counter, ext or ""))}
end

--- Create a symbol from a string or file table. Error if not possible.
function M.symbol(obj, prefix)
    local filename = M.filename(obj)
    if filename == nil then
        error("could not determine filename for symbol")
    end
    local basename = path.splitext(path.basename(filename))
    return (prefix or "") .. wfutil.to_c_identifier(basename)
end

--- Retrieve a filename from a string or file table.
function M.filename(obj)
    if type(obj) == "table" then
        return obj.file
    elseif type(obj) == "string" then
        return obj
    else
        return nil
    end
end

--- Convert a filename or data file to a file table.
function M.to_file(obj)
    if type(obj) == "table" then
        if obj.file ~= nil then
            return obj
        elseif obj.data ~= nil then
            local result = M.tmpfile(".tmp")
            local file <close> = io.open(result.file, "wb")
            file:write(obj.data)
            return result
        end
    elseif type(obj) == "string" then
        return {["file"]=obj}
    else
        error("unsupported type")
    end
end

--- Convert a string or file table to a data table.
function M.to_data(obj)
    if type(obj) == "table" then
        if obj.data ~= nil then
            return obj
        elseif obj.file ~= nil then
            local result = {}
            local file <close> = io.open(obj.file, "rb")
            result.data = file:read("*all")
            return result
        end
    elseif type(obj) == "string" then
        return {["data"]=obj}
    else
        error("unsupported type")
    end
end

--- Access a file without opening or closing it.
-- This is required to correctly emit Makefile dependency files, if a file
-- is not accessed via Lua's "io" package (for example, by an external tool).
-- @tparam string name Filename, as in Lua's "io" package.
-- @tparam string mode File access mode, as in Lua's "io" package.
function M.touch(name, mode)
    _WFPROCESS.access_file(M.filename(name), mode)
end

return M
