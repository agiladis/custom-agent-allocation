package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/agiladis/custom-agent-allocation/internal/config"
	"github.com/agiladis/custom-agent-allocation/internal/service"
	"github.com/redis/go-redis/v9"
)

func RunConsumer(
	ctx context.Context,
	cfg *config.Config,
	rdb *redis.Client,
	assignSvc service.AssignService,
) {
	stream := cfg.RedisStream
	group := cfg.RedisGroup
	consumer := cfg.RedisConsumer

	backoff := 10 * time.Second

	for {
		// 1) Check message (not yet-ack)
		pending, err := rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: stream,
			Group:  group,
			Start:  "-",
			End:    "+",
			Count:  1,
		}).Result()

		var msg redis.XMessage

		if err != nil && err != redis.Nil {
			log.Printf("[consumer] XPENDING error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(pending) > 0 {
			// Claim pending message
			entries, err := rdb.XClaim(ctx, &redis.XClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: consumer,
				MinIdle:  0, // langsung claim
				Messages: []string{pending[0].ID},
			}).Result()
			if err != nil {
				log.Printf("[consumer] XCLAIM error for %s: %v", pending[0].ID, err)
				time.Sleep(time.Second)
				continue
			}
			msg = entries[0]
		} else {
			// Claim new message
			res, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    1,
				Block:    5 * time.Second,
			}).Result()
			if err != nil {
				if err == redis.Nil {
					continue // no new messages
				}
				log.Printf("[consumer] XReadGroup error: %v", err)
				time.Sleep(time.Second)
				continue
			}
			msg = res[0].Messages[0]
		}

		roomVal, ok := msg.Values["room_id"]
		if !ok {
			log.Printf("[consumer] message %s missing room_id", msg.ID)
			_ = rdb.XAck(ctx, stream, group, msg.ID)
			continue
		}
		roomID := fmt.Sprintf("%v", roomVal)

		// assign
		err = assignSvc.AssignCustomer(ctx, roomID)
		if err != nil {
			log.Printf("[consumer] assign failed for room %s: %v — retry in %s", roomID, err, backoff)
			time.Sleep(backoff)

			continue
		}

		// success → ack move next message
		if err := rdb.XAck(ctx, stream, group, msg.ID).Err(); err != nil {
			log.Printf("[consumer] XAck failed for %s: %v", msg.ID, err)
		} else {
			log.Printf("[consumer] room %s assigned successfully", roomID)
		}
	}
}
