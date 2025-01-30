function NewTree(conf) {
    initJstree(conf)

    if (conf.parent) {
        configureEvents(conf)
    }
}

function initJstree(conf) {
    var treeDiv = $("#" + conf.id)
    var overlayDiv = treeDiv.prev()
    var defaultMsgDiv = overlayDiv.children(":first")

    var jstreeConf = {
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
    }

    if (conf.contextmenu) {
        plugins.push("contextmenu")
        jstreeConf.contextmenu = {
            items: conf.contextmenu
        }
    }

    treeDiv.jstree(jstreeConf)

    treeDiv.on("refresh.jstree", function (e, data) {
        if (Object.keys(data.instance._model.data).length == 1) {
          if (selectedId == "") {
            defaultMsgDiv.text("Select a test");
          } else {
            defaultMsgDiv.text("empty");
          }
          overlayDiv.removeClass("d-none");
  
          return;
        }
  
        overlayDiv.addClass("d-none");
        var tree = $(this).jstree(true);

        tree.open_all()
  
        firstFile = Object.values(data.instance._model.data).find((item) => item.type == "file");
        if (firstFile) {
          tree.select_node(firstFile);
        }
    })
    .on("select_node.jstree", function (e, data) {
        if (data.node.type == "file") {
          var model = monaco.editor.createModel(data.node.data, "yaml");
          model.id = data.node.id;
  
          conf.editor.setModel(model);
          conf.editor.updateOptions({ readOnly: conf.readOnly });
        }
    })

    if (conf.contextmenu) {
        treeDiv.on("contextmenu", function (e) {
            if (selectedId == "" || e.target.id != conf.id) {
                return;
            }
            e.preventDefault();

            tree = $(this).jstree(true);
            tree.deselect_all(true);
            obj = tree.get_node("#");
            tree._show_contextmenu(obj, e.pageX, e.pageY, rootMenu());
        })
    }
}