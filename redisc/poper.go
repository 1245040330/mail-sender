package redisc

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/toolkits/pkg/logger"

	"github.com/n9e/mail-sender/dataobj"
)

func Pop(count int, queue string) []*dataobj.Message {
	var lst []*dataobj.Message

	rc := RedisConnPool.Get()
	defer rc.Close()

	for i := 0; i < count; i++ {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				logger.Errorf("rpop queue:%s failed, err: %v", queue, err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var message dataobj.Message
		err = json.Unmarshal([]byte(reply), &message)
		if err != nil {
			logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
			continue
		}

		lst = append(lst, &message)
	}

	return lst
}

/**
添加报警数据
*/
func AddMessage(message *dataobj.Message) (string, error) {
	var lst []*dataobj.Message
	rc := RedisConnPool.Get()
	defer rc.Close()
	//删除
	del, err := redis.Bool(rc.Do("DEL", "alarm-message"))
	if err != nil {
		logger.Errorf("delete message failed, err: %v, redis reply: %v", err, del)
		return "", err
	}
	reply, err := redis.String(rc.Do("GET", "alarm-message"))
	if err != nil {
		if err != redis.ErrNil {
			logger.Errorf("rpop queue:%s failed, err: %v", "alarm-message", err)
		}
	}
	if reply == "" || reply == "nil" {
		//读取到的是空 新增alarm-message数据
		lst = append(lst, message)

	} else {
		err = json.Unmarshal([]byte(reply), &lst)
		if err != nil {
			logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
			return "json转换失败", err
		}
		lst = append(lst, message)
	}
	jsonArr, err := json.Marshal(lst)
	if err != nil {
		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
		return "json转换失败", err
	}

	setres, err := redis.String(rc.Do("SET", "alarm-message", jsonArr))
	//设置数据过期时间为3分钟
	redis.String(rc.Do("EXPIRE", "alarm-message", 180))
	if err != nil {
		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
		return "json转换失败", err
	}
	if setres == "" || setres == "nil" {
		logger.Errorf("返回空")
	}
	return "ok", nil
}

/**
查找报警数据
*/
func FindMessage() (*dataobj.Message, error) {
	var lst []*dataobj.Message
	rc := RedisConnPool.Get()
	defer rc.Close()
	reply, err := redis.String(rc.Do("RPOP", "alarm-message"))
	if err != nil {
		if err != redis.ErrNil {
			logger.Errorf("rpop queue:%s failed, err: %v", "alarm-message", err)
		}
		return nil, err
	}

	if reply == "" || reply == "nil" {
		return nil, err
	}

	var message dataobj.Message
	err = json.Unmarshal([]byte(reply), &message)
	if err != nil {
		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
		return nil, err
	}

	lst = append(lst, &message)

	return &message, nil
	// reply, err := redis.String(rc.Do("GET", "alarm-message"))
	// if err != nil {
	// 	if err != redis.ErrNil {
	// 		logger.Errorf("rpop queue:%s failed, err: %v", "alarm-message", err)
	// 	}
	// 	return nil, err
	// }
	// if reply == "" || reply == "nil" {
	// 	//读取到的是空 新增alarm-message数据
	// 	return nil, err

	// } else {
	// 	err = json.Unmarshal([]byte(reply), &lst)
	// 	if err != nil {
	// 		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
	// 		return nil, err
	// 	}
	// 	if len(lst) <= 0 {
	// 		return nil, err
	// 	}

	// 	res := lst[0]
	// 	lst = lst[1:len(lst)]
	// 	jsonArr, err := json.Marshal(lst)
	// 	if err != nil {
	// 		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
	// 		return nil, err
	// 	}
	// 	//删除
	// 	del, err := redis.Bool(rc.Do("DEL", "alarm-message"))
	// 	if err != nil {
	// 		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
	// 		return nil, err
	// 	}
	// 	fmt.Println(del)
	// 	//设置message
	// 	setres, err := redis.String(rc.Do("SET", "alarm-message", jsonArr))
	// 	if err != nil {
	// 		logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
	// 		return nil, err
	// 	}
	// 	if setres == "" || setres == "nil" {
	// 		logger.Errorf("返回空")
	// 	}
	// 	return res, nil
	// }
}
