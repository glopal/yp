function NewTree(conf) {
    var treeId = "#" + conf.id
    var treeDiv = $(treeId)
    var hasAux = $(`${treeId} + div > .aux-btn`).length === 2
    var stdout = ""
    var err = ""

    treeDiv.jstree({
        core: {
        allow_reselect: true,
        check_callback: true,
        themes: {
            name: "default-dark",
            dots: true,
            icons: true,
            variant: "large",
        },
        data: null,
        },
        types: {
        dir: {},
        file: {
            icon: "text-white bi bi-filetype-yml",
        },
        },
        sort: sortFunc,
        plugins: ["types", "sort", "unique"],
    })
    .on("refresh.jstree", function (e, data) {
      if (Object.keys(data.instance._model.data).length == 1) {
        prev = $(this).prev();
        inner = prev.children(":first");
        if (selectedId == "") {
          inner.text("Run a test");
        } else {
          inner.text("empty");
        }
        prev.removeClass("d-none");

        return;
      }

      $(this).prev().addClass("d-none");
      $(this).jstree("open_all");

      firstFile = Object.values(data.instance._model.data).find((item) => item.type == "file");
      if (firstFile) {
        $(this).jstree().select_node(firstFile);
      }
    })
    .on("select_node.jstree", function (e, data) {
        if (data.node.type == "file") {
            var model = monaco.editor.createModel(data.node.data, "yaml");

            conf.editor.setModel(model);
            conf.editor.updateOptions({ readOnly: true });
        }
    })

    treeDiv.on("update", function(e, data) {
        var tree = $(this).jstree(true);
        tree.settings.core.data = Object.values(data[conf.src]);
        tree.refresh(true, true);
    })

    if (hasAux) {
        var stdoutDiv = $(`${treeId} + div > .aux-btn:first-child`)
        var errDiv = $(`${treeId} + div > .aux-btn:nth-child(2)`)

        stdoutDiv.on("click", function() {
            if(!stdout) return

            var model = monaco.editor.createModel(stdout, "yaml");

            conf.editor.setModel(model);
            conf.editor.updateOptions({ readOnly: true });
        })
        treeDiv.on("update", function(e, data) {
            stdoutDiv.removeClass("aux-selected")
            errDiv.removeClass("aux-selected")

            stdout = data.stdout
            err = data.err

            if (stdout) {
                stdoutDiv.addClass("has-content")
            } else {
                stdoutDiv.removeClass("has-content")
            }
    
            if (err) {
                errDiv.addClass("has-content")
            } else {
                errDiv.removeClass("has-content")
            }
        })
        .on("select_node.jstree", function (e, data) {
            if (data.node.type == "file") {
                $(".aux-btn").removeClass("aux-selected");
            }
        })
        .on("refresh.jstree", function (e, data) {
            var tree = $(this).jstree(true);

            if (tree.get_selected().length > 0) return

            if (stdout) {
                stdoutDiv.trigger("click")
            } else if (stderr) {
                errDiv.trigger("click")
            }
        })
    }



    return {
        setData: function(data) {
            treeDiv.trigger("update", data)
        }
    }
}