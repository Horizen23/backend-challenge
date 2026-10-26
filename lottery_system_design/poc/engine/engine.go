package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SearchResult โครงสร้างผลลัพธ์การค้นหาแบบแบ่งหน้า
type SearchResult struct {
	TotalCount int64
	NextCursor int64
	HasMore    bool
	TicketIDs  []int64
	Latency    time.Duration
}

// LeaseResult โครงสร้างผลการจองตั๋ว
type LeaseResult struct {
	TicketID int64
	Owner    string
	TTL      time.Duration
	Latency  time.Duration
}

// Engine ตัวประมวลผลระบบค้นหาและจองสลากด้วย Redis Bitmaps & Lua
type Engine struct {
	client           *redis.Client
	leaseScript      *redis.Script
	paginationScript *redis.Script
}

// NewEngine สร้าง Engine พร้อมเตรียม Script ล่วงหน้า
func NewEngine(client *redis.Client) *Engine {
	return &Engine{
		client:           client,
		leaseScript:      redis.NewScript(MatchAndLeaseScript),
		paginationScript: redis.NewScript(SearchPaginationScript),
	}
}

// IndexTickets นำสลากเข้าสู่ Redis Positional Bitmaps (60 แถว + available 1 แถว)
func (e *Engine) IndexTickets(ctx context.Context, tickets []string) error {
	pipe := e.client.Pipeline()
	batchSize := 5000

	for i, numStr := range tickets {
		// Available inventory (1 = ว่าง)
		pipe.SetBit(ctx, "bitmap:inventory:available", int64(i), 1)

		// 6 ตำแหน่ง (0..5) กับ ค่าเลข (0..9)
		for pos := 0; pos < 6; pos++ {
			digit := numStr[pos] - '0'
			key := fmt.Sprintf("bitmap:pos:%d:digit:%d", pos, digit)
			pipe.SetBit(ctx, key, int64(i), 1)
		}

		if (i+1)%batchSize == 0 || i == len(tickets)-1 {
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// BuildPatternKeys แปลง Pattern หวย เช่น "****23" เป็นรายการ Redis Keys
func (e *Engine) BuildPatternKeys(pattern string) []string {
	var keys []string
	for pos := 0; pos < len(pattern) && pos < 6; pos++ {
		char := pattern[pos]
		if char >= '0' && char <= '9' {
			keys = append(keys, fmt.Sprintf("bitmap:pos:%d:digit:%c", pos, char))
		}
	}
	return keys
}

// Search ค้นหาสลากตามเงื่อนไข พร้อมทำ Cursor-based Pagination
func (e *Engine) Search(ctx context.Context, pattern string, cursor int64, pageSize int) (*SearchResult, error) {
	keys := e.BuildPatternKeys(pattern)
	start := time.Now()

	resSlice, err := e.paginationScript.Run(ctx, e.client, keys, cursor, pageSize).Slice()
	if err != nil {
		return nil, err
	}

	total := resSlice[0].(int64)
	nextCursor := resSlice[1].(int64)
	hasMore := resSlice[2].(int64) == 1

	var ticketIDs []int64
	for i := 3; i < len(resSlice); i++ {
		ticketIDs = append(ticketIDs, resSlice[i].(int64))
	}

	return &SearchResult{
		TotalCount: total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TicketIDs:  ticketIDs,
		Latency:    time.Since(start),
	}, nil
}

// MatchAndLease ค้นหาและจองสลากใบแรกแบบ Atomic 100% (ป้องกัน Overselling)
func (e *Engine) MatchAndLease(ctx context.Context, pattern string, userID string, ttlSeconds int) (*LeaseResult, error) {
	keys := e.BuildPatternKeys(pattern)
	start := time.Now()

	ticketID, err := e.leaseScript.Run(ctx, e.client, keys, userID, ttlSeconds).Int64()
	if err != nil {
		return nil, err
	}

	if ticketID == -1 {
		return &LeaseResult{TicketID: -1, Latency: time.Since(start)}, nil
	}

	ttl, _ := e.client.TTL(ctx, fmt.Sprintf("lease:ticket:%d", ticketID)).Result()

	return &LeaseResult{
		TicketID: ticketID,
		Owner:    userID,
		TTL:      ttl,
		Latency:  time.Since(start),
	}, nil
}

// ReleaseTicket ปล่อยสลากกลับคืนระบบ (กรณีหมดเวลาจอง หรือยกเลิก)
func (e *Engine) ReleaseTicket(ctx context.Context, ticketID int64) error {
	pipe := e.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("lease:ticket:%d", ticketID))
	pipe.SetBit(ctx, "bitmap:inventory:available", ticketID, 1)
	_, err := pipe.Exec(ctx)
	return err
}

// GetMemoryUsage ตรวจสอบขนาด RAM ที่ใช้เก็บ Available Bitmap
func (e *Engine) GetMemoryUsage(ctx context.Context) (int64, error) {
	return e.client.MemoryUsage(ctx, "bitmap:inventory:available").Result()
}
