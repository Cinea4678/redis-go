local redisApi = require('redisApi')
local resp3 = require('resp3')

function Info()
    return [[
	{
		"name": "getAllKeys",
		"commands": ["getallkeys"]
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
    if command == "getallkeys" then
        local keys = redisApi.getAllKey(db)
        --printTable(keys)
        local resp = "All keys in database:"
        for _, key in ipairs(keys) do
            resp = resp .. "\n" .. key
        end
        local r = resp3.newSimpleString(resp)
        return resp3.toRESP3String(r)
    end

    local defaultResp = resp3.newSimpleError("ERR unknown command")
    return resp3.toRESP3String(defaultResp)
end
