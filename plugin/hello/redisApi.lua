------------------
-- Redis API 接口存根
-- 用于配合LSP提供智能提示
------------------
local redisApi = {}

---设置指定的键值
---@param db integer
---@param key string
---@param val string
function redisApi.setKey(db, key, val) end

---设置指定的键值（数字）
---@param db integer
---@param key string
---@param val integer
function redisApi.setKeyInt(db, key, val) end

---删除指定的键
---@param db integer
---@param key string
function redisApi.deleteKey(db, key) end

---获取指定键的值
---@param db integer
---@param key string
---@return string
function redisApi.getKey(db, key) return "" end

---获取指定键的值（数字）
---@param db integer
---@param key string
---@return integer
function redisApi.getKeyInt(db, key) return 0 end

---设置指定键的过期时间
---@param db integer
---@param key string
---@param expire integer
function redisApi.setExpire(db, key, expire) end

---获取指定键的过期时间
---@param db integer
---@param key string
---@return integer
function redisApi.getExpire(db, key) return 0 end

---获取所有键
---@param db integer
---@return table
function redisApi.getAllKey(db) return {} end

return redisApi
