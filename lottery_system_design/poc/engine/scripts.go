package engine

// MatchAndLeaseScript - Atomic Lua script สำหรับค้นหาตั๋วใบแรกที่ว่างและทำการล็อก (Hold) ทันที
// KEYS[1..N]: Keys ของตำแหน่งตัวเลข เช่น 'bitmap:pos:4:digit:2', 'bitmap:pos:5:digit:3'
// ARGV[1]: User ID
// ARGV[2]: TTL (วินาที)
const MatchAndLeaseScript = `
local user_id = ARGV[1]
local ttl = tonumber(ARGV[2])
local temp_key = "temp:search:" .. user_id
local available_mask = "bitmap:inventory:available"

-- 1. Bitwise AND รวมเงื่อนไขตัวเลขกับสต็อกที่ยังว่างอยู่
local bitop_args = {"AND", temp_key, available_mask}
for i = 1, #KEYS do
    table.insert(bitop_args, KEYS[i])
end
redis.call("BITOP", unpack(bitop_args))

-- 2. ค้นหา Ticket ID แรกที่เป็น 1
local ticket_id = redis.call("BITPOS", temp_key, 1)
redis.call("DEL", temp_key)

if ticket_id == -1 then
    return -1
end

-- 3. Atomic Lock: ตัดตั๋วออกจาก Available Inventory (0 = ไม่ว่าง)
redis.call("SETBIT", available_mask, ticket_id, 0)

-- 4. บันทึกสิทธิ์การจอง พร้อม TTL นับถอยหลัง
redis.call("SET", "lease:ticket:" .. ticket_id, user_id, "EX", ttl)

return ticket_id
`

// SearchPaginationScript - Lua script สำหรับค้นหาแบบ Cursor-based Pagination
// ARGV[1]: cursor (Ticket ID ล่าสุดที่เพิ่งดึงไป, เริ่มแรกส่ง -1)
// ARGV[2]: page_size (จำนวนที่ต้องการต่อหน้า)
// Return: [total_count, next_cursor, has_more, ticket_id_1, ticket_id_2, ...]
const SearchPaginationScript = `
local cursor = tonumber(ARGV[1])
local page_size = tonumber(ARGV[2])
local temp_key = "temp:page:search"
local available_mask = "bitmap:inventory:available"

-- 1. Bitwise AND รวมเงื่อนไขทั้งหมดกับตั๋วที่ว่าง
local bitop_args = {"AND", temp_key, available_mask}
for i = 1, #KEYS do
    table.insert(bitop_args, KEYS[i])
end
redis.call("BITOP", unpack(bitop_args))

-- 2. นับจำนวนตั๋วทั้งหมดที่ตรงเงื่อนไขด้วย BITCOUNT
local total_count = redis.call("BITCOUNT", temp_key)

-- 3. ดึง Ticket ID ตามจำนวน page_size ด้วย BITPOS ต่อจาก cursor
local items = {}
local current_pos = 0
if cursor >= 0 then
    current_pos = cursor + 1
end

for i = 1, page_size do
    local found = redis.call("BITPOS", temp_key, 1, current_pos, -1, "BIT")
    if found == -1 then
        break
    end
    table.insert(items, found)
    current_pos = found + 1
end

-- 4. ตรวจสอบว่ายังมีหน้าถัดไปอีกหรือไม่
local has_more = 0
local next_cursor = -1
if #items > 0 then
    next_cursor = items[#items]
    local check_next = redis.call("BITPOS", temp_key, 1, next_cursor + 1, -1, "BIT")
    if check_next ~= -1 then
        has_more = 1
    end
end

redis.call("DEL", temp_key)

local result = {total_count, next_cursor, has_more}
for _, item in ipairs(items) do
    table.insert(result, item)
end

return result
`
