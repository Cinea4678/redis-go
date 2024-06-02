local redisApi = require('redisApi')
local resp3 = require('resp3')

-- matchkeys功能，用户传入str，返回当前有含有str的key名

function Info()
    return [[
	{
		"name": "matchkey",
		"commands": ["matchkey"]
	}
	]]
end

-- function PrintTable(tbl, indent)
--     if not indent then indent = 0 end
--     for k, v in pairs(tbl) do
--         local formatting = string.rep("  ", indent) .. k .. ": "
--         if type(v) == "table" then
--             print(formatting)
--             PrintTable(v, indent+1)
--         else
--             print(formatting .. tostring(v))
--         end
--     end
-- end

 function Handle(db, reqStr)
     local req = resp3.fromString(reqStr)
     assert(req.t == resp3.typeChars.typeArray, "Error: Invalid Command") -- 确保传入格式正确

     local command = string.lower(req.elems[1].str)
     if command == "matchkey" then
         if #req.elems == 2 then
             local pattern = req.elems[2].str -- 取出参数，用作模式匹配

             -- 输出pattern的类型和内容
             --print("Type of pattern:", type(pattern))
             --print("Content of pattern:", pattern)

             local keysTable = redisApi.getAllKey(db)
             local matchedKeys = {}

             -- 使用lulpeg进行模式匹配
             --local lpeg = require"lulpeg"
             --local p = lpeg.P(pattern)
             --for i, key in ipairs(keysTable) do
             --    if p:match(key) then
             --        table.insert(matchedKeys, key)
             --    end
             --end

             -- 使用Lua原生模式匹配
             for i, key in ipairs(keysTable) do
                 if string.match(key, pattern) then
                     table.insert(matchedKeys, key)
                 end
             end

             -- 创建返回值
             local resp = "Matched Keys: \n" .. table.concat(matchedKeys, "\n")
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

--function Handle(db, reqStr)
--    local req = resp3.fromString(reqStr)
--    assert(req.t == resp3.typeChars.typeArray, "Error: Invalid Command") -- 确保传入格式正确
--
--    local command = string.lower(req.elems[1].str)
--    if command == "matchkey" then
--        local pattern = "k" -- 取出参数，用作模式匹配
--        -- 输出pattern的类型和内容
--        print("Type of pattern:", type(pattern))
--        print("Content of pattern:", pattern)
--        local keysTable = {
--            "apple",
--            "kite",
--            "keyboard",
--            "socket",
--            "book",
--            "glass",
--            "baker",
--            "chocolate",
--            "window",
--        }
--        local matchedKeys = {}
--        -- 使用lulpeg进行模式匹配
--        local lpeg = require "lulpeg"
--        local p = lpeg.P(pattern)
--        for i, key in ipairs(keysTable) do
--            if p:match(key) then
--                table.insert(matchedKeys, key)
--            end
--        end
--        -- 创建返回值
--        local resp = "Matched Keys: " .. table.concat(matchedKeys, ", ")
--        return resp3.toRESP3String(resp)
--    end
--
--    local defaultResp = resp3.newSimpleError("ERR unknown command")
--    return resp3.toRESP3String(defaultResp)
--end
