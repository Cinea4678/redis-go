local lulpeg = require"lulpeg"
local pattern = lulpeg.P("Hello")
local match_result = pattern:match("Hello, world!")
if match_result then
    print("匹配成功: " .. match_result)
else
    print("未匹配。")
end

local match_error = pattern:match("error word")
if match_error then
    print("匹配成功: " .. match_result)
else
    print("未匹配。")
end