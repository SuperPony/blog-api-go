/*
 * Copyright 2021 SuperPony <superponyyy@gmail.com>. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"blog-go/internal/pkg/middleware"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	promethium "github.com/zsais/go-gin-prometheus"
)

type GenericServer struct {
	// SecureServingInfo holds configuration of the TLS server.
	// SecureServingInfo *SecureServingInfo

	// InsecureServingInfo holds configuration of the insecure HTTP server.
	InsecureServingInfo *InsecureServingInfo
	// ShutdownTimeout     time.Duration

	middlewares     []string
	mode            string
	health          bool
	enableMetrics   bool
	enableProfiling bool

	insecureServer *http.Server
	secureServer   *http.Server
	*gin.Engine
}

func initGenericAPIServer(s *GenericServer) {
	s.Setup()
	// s.InstallMiddlewares()
	// s.InstallAPIs()
}

// Setup do some setup work before the service starts
func (s *GenericServer) Setup() {
	gin.SetMode(s.mode)

	// gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	// 	// TODO 换成日志包输出
	// 	fmt.Printf("%-6s %-s --> %s (%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	// }
}

// InstallMiddlewares install global middlewares
func (s *GenericServer) InstallMiddlewares() {
	for _, m := range s.middlewares {
		// MiddlewareName | Middleware-name | Middleware.Name => middleware-name
		formatMw := strings.ToLower(strings.NewReplacer(".", "-", "_", "-").Replace(m))
		mw, ok := middleware.Middlewares[formatMw]
		if !ok {
			log.Fatalf("can not find middleware: %s", m)
		}
		fmt.Printf("installed middleware: %s\n", formatMw)

		s.Use(mw)
	}
}

// InstallAPIs install generic apis
func (s *GenericServer) InstallAPIs() {
	if s.health {
		s.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	// 开启 gin 监控
	if s.enableMetrics {
		prometheus := promethium.NewPrometheus("gin")
		prometheus.Use(s.Engine)
	}

	// 开启性能分析
	if s.enableProfiling {
		pprof.Register(s.Engine)
	}

	// TODO 版本管理功能
}

func (s *GenericServer) Run() error {
	// TODO 排插到服务器无法访问的问题，原因尚未查明
	// Addr: "127.0.0.1:8080" mac 下可以正常访问 centos7 下不行
	// Addr: ":8080" 都可以正常访问
	s.insecureServer = &http.Server{
		Addr:    s.InsecureServingInfo.Address(),
		Handler: s.Engine,
	}

	if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
		return err
	}

	// https
	// s.secureServer = &http.Server{
	// 	Addr:    s.SecureServingInfo.Host(),
	// 	Handler: s,
	// }

	// var eg errgroup.Group
	//
	// eg.Go(func() error {
	// 	if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	// 		log.Fatal(err.Error())
	// 		return err
	// 	}
	//
	// 	return nil
	// })
	//
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	// if s.health {
	// 	if err := s.ping(ctx); err != nil {
	// 		return err
	// 	}
	// }
	//
	// if err := eg.Wait(); err != nil {
	// 	log.Fatal(err.Error())
	// }

	return nil
}

func (s *GenericServer) ping(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/healthz", s.InsecureServingInfo.Host)
	if strings.Contains(s.InsecureServingInfo.Host, "0.0.0.0") {
		url = fmt.Sprintf("http://127.0.0.1:%s/healthz", strings.Split(s.InsecureServingInfo.Host, ":")[1])
	}

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			// TODO 使用日志包记录健康检测
			log.Println("The router has been deployed success")
			_ = resp.Body.Close()
			return nil
		}

		time.Sleep(time.Second)

		select {
		case <-ctx.Done():
			log.Println("can not ping http server.")
		default:
		}
	}
}

// Close graceful shutdown the api server
func (s *GenericServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.secureServer.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown secure server failed: %s", err.Error())
	}

	if err := s.insecureServer.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown insecure secure server failed: %s", err.Error())
	}

}
