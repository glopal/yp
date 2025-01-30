package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/glopal/yp/vfs"
	"github.com/glopal/yp/yplib"
	"github.com/spf13/afero"
)

func main() {
	ts, err := vfs.NewTestSuiteFs("testdata")
	if err != nil {
		panic(err)
	}

	// afero.NewCopyOnWriteFs()

	// memMapFs, err := vfs.ToMemMapFs()
	// if err != nil {
	// 	panic(err)
	// }

	// osFs := afero.NewOsFs()
	// syncFs := afero.NewCopyOnWriteFs(memMapFs, osFs)

	r := gin.Default()

	r.LoadHTMLFiles("playground/index.html") // either individual files like this
	// r.LoadHTMLGlob("index/*")        // or a glob pattern
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	type YpOutput struct {
		Output map[string]vfs.JsTreeNode `json:"output"`
		Stdout string                    `json:"stdout"`
		Err    string                    `json:"err"`
	}
	r.GET("/run", func(c *gin.Context) {
		id := c.Query("id")

		test, exists := ts.Get(id)
		if !exists {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		ofs := afero.NewMemMapFs()
		b := bytes.NewBuffer([]byte{})

		ypErr := ""
		err := yplib.WithOptions(yplib.WithFS(afero.NewIOFS(test.Input.Fs)), yplib.WithOutputFS(ofs), yplib.WithWriter(b)).Load(".").Out()
		if err != nil {
			ypErr = err.Error()
		}

		vofs, err := vfs.UnmarshalFs(afero.NewIOFS(ofs))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, YpOutput{
			Output: vofs.ToJsTreeMap(),
			Stdout: b.String(),
			Err:    ypErr,
		})
	})

	r.POST("/update", func(ctx *gin.Context) {
		var updateBody UpdateBody
		if err := ctx.BindJSON(&updateBody); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := updateBody.Update(ts); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusOK)
	})

	r.POST("/update/test", func(ctx *gin.Context) {
		var updateTestBody UpdateTestBody
		if err := ctx.BindJSON(&updateTestBody); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := updateTestBody.Update(ts); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusOK)
	})

	// r.POST("/test/update", func(ctx *gin.Context) {
	// 	var update UpdateBody
	// 	if err := ctx.BindJSON(&update); err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	vfd, exists := vfs.Get(update.TestId)
	// 	if !exists {
	// 		ctx.Status(http.StatusNotFound)
	// 		return
	// 	}

	// 	subfs := yamlp.NewVFS()
	// 	if err := subfs.UnmarshalYamlString(*vfd.Content); err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	_, exists = subfs.Set(update.FileId, &update.Content)
	// 	if !exists {
	// 		ctx.Status(http.StatusNotFound)
	// 		return
	// 	}

	// 	data, err := subfs.ToYaml()
	// 	if err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	if err := os.WriteFile(update.TestId, data, 0755); err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	ctx.Status(http.StatusOK)
	// })
	// r.GET("/test", func(ctx *gin.Context) {
	// 	path := ctx.Query("path")
	// 	vfs := yamlp.NewVFS()

	// 	err := vfs.UnmarshalYaml(path)
	// 	if err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	tree, err := vfs.ToJtree()
	// 	if err != nil {
	// 		ctx.AbortWithError(http.StatusInternalServerError, err)
	// 		return
	// 	}

	// 	ctx.JSON(http.StatusOK, tree)
	// })
	r.GET("/tests.json", func(ctx *gin.Context) {
		data, err := json.Marshal(ts)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})
	r.Use(static.Serve("/", static.LocalFile("playground", false)))
	// r.Use(static.Serve("/", static.LocalFile("./playground/index.html", true)))
	// r.Use(static.ServeRoot("/assets", "./playground/assets"))
	// r.NotFound(static.Serve("/public"))

	// r.Static("/", "./playground")

	// Listen and serve on 0.0.0.0:8080
	r.Run("127.0.0.1:8080")
}

// func syncer(vfs yamlp.VFS)
