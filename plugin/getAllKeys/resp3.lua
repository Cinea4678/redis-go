-- Lua-RESP3
--
-- Author: Cinea (Zhang Yao)
-- License: Apache
--
-- Usage:
--
--
-- declare module
local resp3 = {}

-- declare const values
local CRLF = "\r\n"
local streamMarkerPrefix = "$EOF:"

-- resp3 type char

resp3.typeChars = {}

-- simple types

resp3.typeChars.typeBlobString = "$" -- $<length>\r\n<bytes>\r\n
resp3.typeChars.typeSimpleString = "+" -- +<string>\r\n
resp3.typeChars.typeSimpleError = "-" -- -<string>\r\n
resp3.typeChars.typeNumber = ":" -- :<number>\r\n
resp3.typeChars.typeNull = "_" -- _\r\n
resp3.typeChars.typeDouble = "," -- ,<floating-point-number>\r\n
resp3.typeChars.typeBoolean = "#" -- #t\r\n or #f\r\n
resp3.typeChars.typeBlobError = "!" -- !<length>\r\n<bytes>\r\n
resp3.typeChars.typeVerbatimString = "=" -- =<length>\r\n<format(3 bytes):><bytes>\r\n
resp3.typeChars.typeBigNumber = "(" -- (<big number>\n

-- aggregate data types

resp3.typeChars.typeArray = "*" -- *<elements number>\r\n...
resp3.typeChars.typeMap = "%" -- %<elements number>\r\n...
resp3.typeChars.typeSet = "~" -- ~<elements number>\r\n...
resp3.typeChars.typeAttribute = "|" -- |~<elements number>\r\n...
resp3.typeChars.typePush = ">" -- ><elements number>\r\n<first item is String>\r\n...

-- special type

resp3.typeChars.typeStream = "$EOF:" -- $EOF:<40 bytes marker><CR><LF>...

-- functions

function resp3.toRESP3String(r)
    local buf = ""

    if r.attrs ~= nil and #r.attrs > 0 then
        buf = buf .. resp3.typeChars.typeAttribute .. #r.attrs .. CRLF

        for key, val in pairs(r.attrs) do
            buf = buf .. key.t .. key:toRESP3String() .. val.t ..
                      val:toRESP3String()
        end
    end

    buf = buf .. r.t .. r:toRESP3String()

    return buf
end

-------------------
-- type's table
-------------------

-- BlobString

local BlobString = {}

---@return string
function BlobString:toRESP3String() return #self.str .. CRLF .. self.str .. CRLF end

function resp3.newBlobString(s)
    local v = {t = resp3.typeChars.typeBlobString, str = s}
    setmetatable(v, {__index = BlobString})
    return v
end

-- SimpleString

local SimpleString = {}

---@return string
function SimpleString:toRESP3String() return self.str .. CRLF end

function resp3.newSimpleString(s)
    local v = {t = resp3.typeChars.typeSimpleString, str = s}
    setmetatable(v, {__index = SimpleString})
    return v
end

-- VerbatimString

local VerbatimString = {}

---@return string
function VerbatimString:toRESP3String()
    return (#self.str + 4) .. CRLF .. self.strFmt .. ":" .. self.str .. CRLF
end

function resp3.newVerbatimString(s, fmt)
    local v = {t = resp3.typeChars.typeVerbatimString, str = s, strFmt = fmt}
    setmetatable(v, {__index = VerbatimString})
    return v
end

-- SimpleError

local SimpleError = {}

---@return string
function SimpleError:toRESP3String() return self.err .. CRLF end

function resp3.newSimpleError(err)
    local v = {t = resp3.typeChars.typeSimpleError, err = err}
    setmetatable(v, {__index = SimpleError})
    return v
end

-- BlobError

local BlobError = {}

---@return string
function BlobError:toRESP3String() return #self.err .. CRLF .. self.err .. CRLF end

function resp3.newBlobError(err)
    local v = {t = resp3.typeChars.typeBlobError, err = err}
    setmetatable(v, {__index = BlobError})
    return v
end

-- Number

local Number = {}

---@return string
function Number:toRESP3String() return self.num .. CRLF end

function resp3.newNumber(n)
    local v = {t = resp3.typeChars.typeNumber, num = n}
    setmetatable(v, {__index = Number})
    return v
end

-- Double

local Double = {}

---@return string
function Double:toRESP3String() return self.num .. CRLF end

function resp3.newDouble(n)
    local v = {t = resp3.typeChars.typeDouble, num = n}
    setmetatable(v, {__index = Double})
    return v
end

-- BigNumber

local BigNumber = {}

---@return string
function BigNumber:toRESP3String() return self.num .. CRLF end

function resp3.newBigNumber(n)
    local v = {t = resp3.typeChars.typeBigNumber, num = n}
    setmetatable(v, {__index = BigNumber})
    return v
end

-- Null

local Null = {}

---@return string
function Null:toRESP3String() return CRLF end

function resp3.newNull()
    local v = {t = resp3.typeChars.typeNull}
    setmetatable(v, {__index = Null})
    return v
end

-- Boolean

local Boolean = {}

---@return string
function Boolean:toRESP3String()
    if self.bool then
        return "t" .. CRLF
    else
        return "f" .. CRLF
    end
end

function resp3.newBoolean(b)
    local v = {t = resp3.typeChars.typeBoolean, bool = b}
    setmetatable(v, {__index = Boolean})
    return v
end

-- List

local List = {}

---@return string
function List:toRESP3String()
    local buf = #self.elems .. CRLF
    for _, v in ipairs(self.elems) do buf = v.t .. v:toRESP3String() end
    return buf .. CRLF
end

function resp3.newArray(e)
    local v = {t = resp3.typeChars.typeArray, elems = e}
    setmetatable(v, {__index = List})
    return v
end

function resp3.newSet(e)
    local v = {t = resp3.typeChars.typeSet, elems = e}
    setmetatable(v, {__index = List})
    return v
end

function resp3.newPush(e)
    local v = {t = resp3.typeChars.typePush, elems = e}
    setmetatable(v, {__index = List})
    return v
end

-- Map

local Map = {}

---@return string
function Map:toRESP3String()
    local buf = #self.kv .. CRLF
    for k, v in pairs(self.kv) do
        buf = k.t .. k:toRESP3String() .. v.t .. v:toRESP3String() .. CRLF
    end
    return buf .. CRLF
end

function resp3.newMap(kv)
    local v = {t = resp3.typeChars.typeMap, kv = kv}
    setmetatable(v, {__index = Map})
    return v
end

----------------
-- parser
----------------

local Reader = {buf = "", cur = 1}

---Parses a RESP3 value.
---@param s string
function resp3.fromString(s)
    local v = {buf = s, cur = 1}
    setmetatable(v, {__index = Reader})
    return v:readValue()
end

function Reader:readValue()
    local line = self:readLine()
    assert(line ~= nil and #line >= 3, "Error: Invalid Syntax")

    local attrs = {}
    if line:sub(1, 1) == resp3.typeChars.typeAttribute then
        attrs = self:readAttr(line)
        line = self:readLine()
    end
    assert(line ~= nil, "Error: Invalid Syntax")

    if line:sub(1, 1) == resp3.typeChars.typeBlobString and #line == 45 and
        line:sub(1, 5) == streamMarkerPrefix then return nil, line:sub(1, 5) end

    local v = {t = line:sub(1, 1), attrs = attrs}

    if v.t == resp3.typeChars.typeSimpleString then
        v.str = line:sub(2, line:len() - 2)
    elseif v.t == resp3.typeChars.typeBlobString then
        v.str = self:readBlobString(line)
    elseif v.t == resp3.typeChars.typeVerbatimString then
        local s = self:readBlobString(line)
        assert(#s >= 4, "Error: Invalid Syntax")
        v.str = s:sub(5)
        v.strFmt = s:sub(1, 3)
    elseif v.t == resp3.typeChars.typeSimpleError then
        v.err = line:sub(2, line:len() - 2)
    elseif v.t == resp3.typeChars.typeBlobError then
        v.err = self:readBlobString(line)
    elseif v.t == resp3.typeChars.typeNumber then
        v.num = self:readNumber(line)
    elseif v.t == resp3.typeChars.typeBigNumber then
        v.num = self:readBigInt(line)
    elseif v.t == resp3.typeChars.typeDouble then
        v.num = self:readDouble(line)
    elseif v.t == resp3.typeChars.typeNull then
        assert(#line == 3, "Error: Invalid Syntax")
    elseif v.t == resp3.typeChars.typeBoolean then
        v.bool = self:readBoolean(line)
    elseif v.t == resp3.typeChars.typeArray or v.t == resp3.typeChars.typeSet or
        v.t == resp3.typeChars.typePush then
        v.elems = self:readArray(line)
    elseif v.t == resp3.typeChars.typeMap then
        v.kv = self:readMap(line)
    end

    return v, nil
end

---@return string|nil
function Reader:readLine()
    local lineIndex = self.buf:find('\n', self.cur)
    if lineIndex == nil then return nil end

    local line = self.buf:sub(self.cur, lineIndex)
    assert(#line > 1 and line:sub(#line - 1, #line - 1) == '\r',
           "Error: Invalid Syntax")
    self.cur = lineIndex + 1
    return line
end

---@param line string
function Reader:getCount(line)
    local endIndex = line:find('\r')
    assert(endIndex ~= nil, "Error: Invalid Syntax")
    return tonumber(line:sub(2, endIndex - 1))
end

---@param line string
function Reader:readBlobString(line)
    local count = self:getCount(line)
    assert(count >= 0, "Error: Invalid Syntax")

    local buf = self.buf:sub(self.cur, self.cur + count + 2)
    self.cur = self.cur + count + 2
    return buf:sub(1, count)
end

---@param line string
function Reader:readNumber(line)
    local v = line:sub(2, #line - 2)
    return tonumber(v)
end

---@param line string
function Reader:readDouble(line)
    local v = line:sub(2, #line - 2)
    if v == "inf" then
        return math.huge
    elseif v == "-inf" then
        return -math.huge
    end
    return tonumber(v)
end

---@param line string
function Reader:readBigInt(line)
    local v = line:sub(2, #line - 2)
    return v
end

---@param line string
function Reader:readBoolean(line)
    local v = line:sub(2, #line - 2)
    assert(v == 't' or v == 'f', "Error: Invalid Syntax")
    if v == 't' then
        return true
    else
        return false
    end
end

---@param line string
function Reader:readArray(line)
    local count = self:getCount(line)
    local rt = {}
    for _ = 1, count do
        local v, smp = self:readValue()
        self:isError(v, smp)
        table.insert(rt, v)
    end
    return rt
end

---@param line string
function Reader:readMap(line)
    local count = self:getCount(line)
    local rt = {}
    for _ = 1, count do
        local k, smp1 = self:readValue()
        self:isError(k, smp1)
        local v, smp2 = self:readValue()
        self:isError(v, smp2)
        assert(k ~= nil, "Error: Invalid Syntax")
        rt[k] = v
    end
    return rt
end

---@param line string
function Reader:readAttr(line) return self:readMap(line) end

---@param smp string|nil
function Reader:isError(val, smp)
    assert(val ~= nil and (smp == nil or #smp == 0), "Error: Invalid Syntax")
end

return resp3
