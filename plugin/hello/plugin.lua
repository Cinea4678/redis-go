local redisApi = require('redisApi')
local resp3 = require('resp3')

function Info()
    return [[
	{
		"name": "Hello",
		"commands": ["hellolua"]
	}
	]]
end

function Handle(db, reqStr)
    local req = resp3.fromString(reqStr)
    assert(req.t == resp3.typeChars.typeArray, "Error: Invalid Command") -- 确保传入格式正确

    local command = string.lower(req.elems[1].str)
    if command == "hellolua" then
        if #req.elems > 1 then
            local name = req.elems[2].str -- 取出参数
            local resp = "Hello, " .. name .. "! I am a lua extension."
            local r = resp3.newSimpleString(resp)
            return resp3.toRESP3String(r)
        else
            local keyName = "hello-lua-count"
            local count = redisApi.getKeyInt(db, keyName)
            local resp = ""
            if count == nil or count == 0 then
                resp = resp .. "Hi! It's your first time to call me."
                redisApi.setKeyInt(db, keyName, 1)
            else
                resp = resp .. "Hi again! You have called me for " .. count ..
                           " times."
                redisApi.setKeyInt(db, keyName, count + 1)
            end
            local r = resp3.newSimpleString(resp)
            return resp3.toRESP3String(r)
        end
    end

    local defaultResp = resp3.newSimpleError("ERR unknown command")
    return resp3.toRESP3String(defaultResp)
end
