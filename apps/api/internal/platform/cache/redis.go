package cache

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)
func Open(address,password string,db int)(*redis.Client,error){ client:=redis.NewClient(&redis.Options{Addr:address,Password:password,DB:db,DialTimeout:5*time.Second,ReadTimeout:3*time.Second,WriteTimeout:3*time.Second}); ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second); defer cancel(); if err:=client.Ping(ctx).Err();err!=nil{client.Close();return nil,fmt.Errorf("ping redis: %w",err)}; return client,nil }
