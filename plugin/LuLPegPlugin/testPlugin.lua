local redisApi = require('redisApi')
local resp3 = require('resp3')

-- matchkey功能，用户传入str，返回当前有含有str的key名

function Info()
    return [[
	{
		"name": "match",
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

function Handle()
    local pattern = "k" -- 取出参数，用作模式匹配
    -- 输出pattern的类型和内容
    print("Type of pattern:", type(pattern))
    print("Content of pattern:", pattern)
    local keysTable = {
        "apple",
        "kite",
        "keyboard",
        "socket",
        "book",
        "glass",
        "baker",
        "chocolate",
        "window",
    }
    local matchedKeys = {}
    -- 使用lulpeg进行模式匹配
    local lpeg = require"lulpeg"
    local p = lpeg.P(pattern)
    for i, key in ipairs(keysTable) do
        if p:match(key) then
            table.insert(matchedKeys, key)
        end
    end
    -- 创建返回值
    local resp = "Matched Keys: " .. table.concat(matchedKeys, ", ")
    print(resp)
end

Handle()