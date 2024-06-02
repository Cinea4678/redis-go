local redisApi = require('redisApi')
local resp3 = require('resp3')
local lulpeg = require('lulpeg')

-- matchkey功能，用户传入str，返回当前有含有str的key名

function Info()
    return [[
	{
		"name": "lulupeg",
		"commands": ["matchkey"]
	}
	]]
end

function printTable(tbl, indent)
    if not indent then indent = 0 end
    for k, v in pairs(tbl) do
        local formatting = string.rep("  ", indent) .. k .. ": "
        if type(v) == "table" then
            print(formatting)
            printTable(v, indent+1)
        else
            print(formatting .. tostring(v))
        end
    end
end

function Handle(db, reqStr)
    local req = resp3.fromString(reqStr)
    assert(req.t == resp3.typeChars.typeArray, "Error: Invalid Command") -- 确保传入格式正确

    local command = string.lower(req.elems[1].str)
    if command == "matchkey" then
        if #req.elems == 2 then
            local name = req.elems[2].str -- 取出参数
            local resp = "Hello, " .. name .. "! I am a lua extension."
            local r = resp3.newSimpleString(resp)
            return resp3.toRESP3String(r)
        else
            local resp = "(Err) Only one parameter is allowed."
            local r = resp3.newSimpleString(resp)
            return resp3.toRESP3String(r)
        end
    end

    local defaultResp = resp3.newSimpleError("ERR unknown command")
    return resp3.toRESP3String(defaultResp)
end
