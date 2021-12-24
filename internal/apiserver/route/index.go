/*
 * Copyright 2021 SuperPony <superponyyy@gmail.com>. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package route

import (
	"blog-api/internal/apiserver/controller/v1/index"
	"github.com/gin-gonic/gin"
)

func Index(e *gin.Engine) {
	v1 := e.Group(V1)
	{
		indexV1 := v1.Group("/index")
		{
			indexController := index.NewIndexController()
			indexV1.GET("/", indexController.Index)
		}
	}
}
