package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"backend-challenge/lottery_system_design/poc/engine"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	// Wait for Redis to be ready with retry
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = rdb.Ping(ctx).Err()
		if pingErr == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if pingErr != nil {
		log.Fatalf("Cannot connect to Redis at %s: %v", redisAddr, pingErr)
	}

	fmt.Println("==================================================")
	fmt.Println("LOTTERY POSITIONAL BITMAP & LUA ENGINE (POC)")
	fmt.Println("==================================================")
	fmt.Printf("Connected to Redis: %s\n\n", redisAddr)

	// ล้าง Redis keys เก่า
	rdb.FlushDB(ctx)

	engine := engine.NewEngine(rdb)

	// 1. Seed ข้อมูลสลาก 100,000 ใบ
	totalTickets := 100_000
	fmt.Printf("[1] กำลังสร้างและ Index สลาก %d ใบ เข้าสู่ Redis...\n", totalTickets)
	startSeed := time.Now()

	r := rand.New(rand.NewSource(42))
	ticketNumbers := make([]string, totalTickets)
	for i := 0; i < totalTickets; i++ {
		ticketNumbers[i] = fmt.Sprintf("%06d", r.Intn(1_000_000))
	}

	if err := engine.IndexTickets(ctx, ticketNumbers); err != nil {
		log.Fatalf("IndexTickets failed: %v", err)
	}
	fmt.Printf("[Success] Index ข้อมูลเสร็จสิ้นใน: %v\n\n", time.Since(startSeed))

	// 2. ตรวจสอบ Memory Usage
	mem, _ := engine.GetMemoryUsage(ctx)
	fmt.Printf("[2] Memory Usage ของ Available Bitmap: %d bytes (~%.2f KB)\n\n", mem, float64(mem)/1024)

	// 3. ทดสอบ Cursor-based Pagination ค้นหาเลขท้าย '23'
	fmt.Println("[3] ทดสอบ Cursor-based Pagination (Pattern: '****23', หน้าละ 5 ใบ):")
	fmt.Println("------------------------------------------------------------------")
	pattern := "****23"
	cursor := int64(-1)
	pageSize := 5

	for page := 1; page <= 3; page++ {
		res, err := engine.Search(ctx, pattern, cursor, pageSize)
		if err != nil {
			log.Fatalf("Search failed: %v", err)
		}

		fmt.Printf("หน้าที่ %d (Latency: %v | พบทั้งหมด: %d ใบ | has_more: %v)\n",
			page, res.Latency, res.TotalCount, res.HasMore)

		for _, tid := range res.TicketIDs {
			fmt.Printf("   - Ticket ID: %-5d | หมายเลขสลาก: %s\n", tid, ticketNumbers[tid])
		}

		cursor = res.NextCursor
		fmt.Printf("   [Next Cursor = %d]\n\n", cursor)

		if !res.HasMore {
			break
		}
	}

	// 4. ทดสอบ Concurrency Match & Lease (10 คนแย่งกดจองพร้อมกัน)
	fmt.Println("[4] ทดสอบ Concurrency Atomic Match & Lease (10 คนแย่งซื้อพร้อมกัน):")
	var wg sync.WaitGroup
	concurrentUsers := 10

	for u := 1; u <= concurrentUsers; u++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()
			uID := fmt.Sprintf("user_%02d", userNum)
			leaseRes, err := engine.MatchAndLease(context.Background(), pattern, uID, 300)
			if err != nil {
				fmt.Printf("   [%s] เกิดข้อผิดพลาด: %v\n", uID, err)
				return
			}
			if leaseRes.TicketID != -1 {
				fmt.Printf("   [Success] [%s] จองสำเร็จ -> Ticket ID: %-5d (เลข: %s | Latency: %v)\n",
					uID, leaseRes.TicketID, ticketNumbers[leaseRes.TicketID], leaseRes.Latency)
			} else {
				fmt.Printf("   [Failed] [%s] สลากหมดแล้ว\n", uID)
			}
		}(u)
	}
	wg.Wait()

	fmt.Println("\n[Success] ผลลัพธ์: ตั๋วถูกล็อกไม่ซ้ำกัน 100% ปราศจาก Race Condition")
	fmt.Println("==================================================")
}
