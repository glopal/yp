package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/glopal/yamlplus/yamlp"
)

func main() {
	vfs := yamlp.NewVFS()

	// afero.NewCopyOnWriteFs()

	err := vfs.UnmarshalDir("tests", 0)
	if err != nil {
		panic(err)
	}

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

	r.POST("/sync", func(ctx *gin.Context) {
		var tree yamlp.Jtree
		if err := ctx.BindJSON(&tree); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// nfs, err := tree.ToVFS()
		// if err != nil {
		// 	ctx.AbortWithError(http.StatusInternalServerError, err)
		// 	return
		// }

		ctx.Status(http.StatusOK)
	})

	r.POST("/create", func(ctx *gin.Context) {
		var node yamlp.JtreeNode
		if err := ctx.BindJSON(&node); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// if _, ok := vfs.Push(node.Id, node.Content); !ok {
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }

		// fsrc, err := vfs.ToMemMapFs()
		// if err != nil {
		// 	ctx.AbortWithError(http.StatusInternalServerError, err)
		// 	return
		// }

		// sync.Sync(context.TODO(), os.DirFS("tests"), fsrc, true)

		// if err := syncFs.Mkdir(node.Id, 0755); err != nil {
		// 	fmt.Println(err)
		// 	ctx.AbortWithError(http.StatusInternalServerError, err)
		// 	return
		// }

		ctx.Status(http.StatusOK)
	})

	type UpdateBody struct {
		TestId  string `json:"testId"`
		FileId  string `json:"fileId"`
		Content string `json:"content"`
	}
	r.POST("/test/update", func(ctx *gin.Context) {
		var update UpdateBody
		if err := ctx.BindJSON(&update); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		vfd, exists := vfs.Get(update.TestId)
		if !exists {
			ctx.Status(http.StatusNotFound)
			return
		}

		subfs := yamlp.NewVFS()
		if err := subfs.UnmarshalYamlString(*vfd.Content); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		_, exists = subfs.Set(update.FileId, &update.Content)
		if !exists {
			ctx.Status(http.StatusNotFound)
			return
		}

		data, err := subfs.ToYaml()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := os.WriteFile(update.TestId, data, 0755); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusOK)
	})
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
		tree, err := vfs.ToJtree()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for _, node := range tree {
			if node.Type == "file" {
				node.Text = strings.TrimSuffix(node.Text, ".yml")
			}
		}

		tree = tree.ContentToTree()

		ctx.JSON(http.StatusOK, tree)
	})
	r.Use(static.Serve("/", static.LocalFile("playground", false)))
	// r.Use(static.Serve("/", static.LocalFile("./playground/index.html", true)))
	// r.Use(static.ServeRoot("/assets", "./playground/assets"))
	// r.NotFound(static.Serve("/public"))

	// r.Static("/", "./playground")

	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")
}

// func syncer(vfs yamlp.VFS)
