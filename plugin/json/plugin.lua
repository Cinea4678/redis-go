local redisApi = require('redisApi')
local resp3 = require('resp3')
local jmespath = require("jmespath")
local json = require("json")

local function jget(db, req)
    if #req.elems < 3 then
        return resp3.toRESP3String(resp3.newSimpleError(
                                       "ERR not enough arguments"))
    end

    local key = req.elems[2].str
    local path = req.elems[3].str
    local jsonRaw = redisApi.getKey(db, key)
    if jsonRaw == nil or jsonRaw == "" then
        return resp3.toRESP3String(resp3.newSimpleError("ERR key not exists"))
    end

    local json_ = json.decode(jsonRaw)
    local result = jmespath.search(path, json_)
    local resp = json.encode(result)
    local r = resp3.newBlobString(resp)
    return resp3.toRESP3String(r)
end

function Info()
    return [[
	{
		"name": "Json",
		"commands": ["jget"]
	}
	]]
end

function Handle(db, reqStr)
    local req = resp3.fromString(reqStr)
    assert(req.t == resp3.typeChars.typeArray, "Error: Invalid Command") -- 确保传入格式正确

    local command = string.lower(req.elems[1].str)
    if command == "jget" then return jget(db, req) end
    local defaultResp = resp3.newSimpleError("ERR unknown command")
    return resp3.toRESP3String(defaultResp)
end

print("\nHERE:", jmespath.search("a", {["a"] = 2}))
