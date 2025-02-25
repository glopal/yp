function NewTree(conf) {
    var hasAux = $(`#${conf.id} + div > .aux-btn`).length === 2
    conf.readOnly = Boolean(conf.readOnly)

    initJstree(conf)

    if (conf.parent) {
        configureEvents(conf)

        $("#" + conf.parent).on("select_change", function (e, data) {
            var tree = $("#" + conf.id).jstree(true);
            tree.settings.core.data = data.node.type == "file" ? Object.values(data.node.data[conf.src]) : null;
            tree.refresh(true, true);

            clearEditor(conf.editor);
        });
    }

    if (hasAux) {
        configureAux(conf)
    }

    if (conf.parent && hasAux) {
        $("#" + conf.id).on("update", function (e, data) {
            var pnode = $("#" + conf.parent).jstree(true).get_node(selectedId);
            pnode.data.stdout = data.stdout;
            pnode.data.err = data.err;
        })
    }

    return {
        id: conf.id,
        $tree: $("#" + conf.id),
        $stdout: $(`#${this.id} + div > .aux-btn:first-child`),
        $err: $(`#${this.id} div > .aux-btn:nth-child(2)`),
        setData: function (data) {
            this.$tree.trigger("update", data)
        },
        getStdout: function () {
            return $(`#${this.id} + div > .aux-btn:first-child`).data("stdout")
        },
        getErr: function () {
            return $(`#${this.id} div > .aux-btn:nth-child(2)`).data("err")
        },
        clear: function () {
            var tree = this.$tree.jstree(true);
            tree.settings.core.data = null;
            tree.refresh(true, true);

            updateAuxData($(`#${this.id} + div > .aux-btn:first-child`), $(`#${this.id} div > .aux-btn:nth-child(2)`), "", "")
        }
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
        jstreeConf.plugins.push("contextmenu")
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
                model.uri.authority = conf.id
                model.data = data

                conf.editor.setModel(model);
                conf.editor.updateOptions({ readOnly: conf.readOnly });
            }
        })
        .on("update", function (e, data) {
            var tree = $(this).jstree(true);

            if (conf.parent) {
                $(this).on("refresh.jstree.once", function (e, data) {
                    var pnode = $("#" + conf.parent).jstree(true).get_node(selectedId)
                    pnode.data[conf.src] = extractChildData(tree);

                    $(this).off("refresh.jstree.once");
                });
            }
            tree.settings.core.data = data[conf.src];
            tree.refresh(true, true);
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

    conf.editor.onDidChangeModel((e) => {
        console.log("onDidChangeModel", conf.id, conf.editor, e)
        if (e.newModelUrl.path == "/empty") {
            console.log("IS EMPTY")
            return
        }

        var model = conf.editor.getModel();

        if (model.uri.authority != conf.id) {
            treeDiv.jstree(true).deselect_all(true)
        }

    })
}

function configureEvents(conf) {
    var childId = "#" + conf.id;
    var parentId = "#" + conf.parent;
    var events = NewNodeListener(conf.id, "")

    events.onUpdate = function (data, parent) {
        var pnode = $(parentId).jstree(true).get_node(parent)
        var op, target, id, content

        if (data.node) {
            var picked = (({ id, data, parent, text, type }) => ({ id, data, parent, text, type }))(data.node);

            pnode.data[conf.src][data.node.id] = picked;

            op = data.node.type == "dir" ? "PUSH_DIR" : "PUSH";
            id = data.node.id
            target = conf.src
            content = data.node.data
        } else {
            pnode.data[data.id] = data.val

            op = "PUSH"
            target = data.id
            content = data.val
        }

        updateTest({
            op: op,
            id: id,
            parentId: parent,
            target: target,
            content: content,
        });
    }

    events.onRename = function (data, parent) {
        var pnode = $(parentId).jstree(true).get_node(parent)
        pnode.data[conf.src] = extractChildData($(childId).jstree(true))
        updateTest({
            op: "RENAME",
            parentId: parent,
            target: conf.src,
            oldId: data.old,
            id: data.node.id,
        });
    }

    events.onDelete = function (data, parent) {
        var pnode = $(parentId).jstree(true).get_node(parent)
        var tree = $(childId).jstree(true)

        delete pnode.data[conf.src][data.node.id]

        data.node.children_d.forEach(child => {
            delete pnode.data[conf.src][child]
        });

        tree.settings.core.data = Object.values(pnode.data[conf.src]);
        tree.refresh(true);

        updateTest({
            op: "DELETE",
            parentId: parent,
            target: conf.src,
            id: data.node.id,
        });
    };

    conf.editor.onDidChangeModel((e) => {
        // editor was cleared and set to the empty model
        // no need to attach update handler, just return
        if (e.newModelUrl.path == "/empty") {
            return
        }
        var model = conf.editor.getModel();
        var updateContent

        if (e.newModelUrl.fragment == "#aux") {
            updateContent = debounce(function () {
                var value = this.model.getValue()

                this.data.data(model.id, value)

                this.events.onUpdate({
                    id: model.id,
                    val: value
                }, this.parent.id)
            }.bind({
                data: model.data,
                parent: $("#" + conf.parent).jstree(true).get_node(selectedId),
                model: model,
                events: events,
            }));
        } else {
            updateContent = debounce(function () {
                var value = this.model.getValue()
                this.data.node.data = value

                this.events.onUpdate(this.data, this.parent.id)
            }.bind({
                data: model.data,
                parent: $(parentId).jstree(true).get_node(selectedId),
                src: conf.src,
                model: model,
                events: events,
            }));
        }

        model.onDidChangeContent(() => {
            updateContent()
        });
    });

    return events
}

function configureAux(conf) {
    var treeId = "#" + conf.id
    var stdoutDiv = $(`${treeId} + div > .aux-btn:first-child`)
    var errDiv = $(`${treeId} + div > .aux-btn:nth-child(2)`)

    if (conf.parent) {
        $("#" + conf.parent).on("select_change", function (e, data) {
            var stdout = data.node.type == 'file' && data.node.data.stdout ? data.node.data.stdout : "";
            var err = data.node.type == 'file' && data.node.data.err ? data.node.data.err : "";
            updateAuxData(stdoutDiv, errDiv, stdout, err)
        });
    }

    stdoutDiv.on("click", function () {
        var stdout = stdoutDiv.data("stdout")
        if (!stdout) return

        // TODO better place to manage this?
        $(treeId).jstree(true).deselect_all(true);

        var model = monaco.editor.createModel(stdout, "yaml");
        model.id = "stdout"
        model.uri.authority = conf.id
        model.uri.fragment = "#aux"
        model.data = stdoutDiv

        conf.editor.setModel(model);
        conf.editor.updateOptions({ readOnly: conf.readOnly });
    })
    errDiv.on("click", function () {
        var err = errDiv.data("err")
        if (!err) return

        // TODO better place to manage this?
        $(treeId).jstree(true).deselect_all(true);
        var model = monaco.editor.createModel(err, "plain");
        model.id = "err"
        model.uri.authority = conf.id
        model.uri.fragment = "#aux"
        model.data = errDiv

        conf.editor.setModel(model);
        conf.editor.updateOptions({ readOnly: conf.readOnly });
    })
    $(treeId).on("update", function (e, data) {
        updateAuxData(stdoutDiv, errDiv, data.stdout, data.err)
    })
        .on("select_node.jstree", function (e, data) {
            if (data.node.type == "file") {
                $(".aux-btn").removeClass("aux-selected");
            }
        })
        .on("refresh.jstree", function (e, data) {
            var tree = $(this).jstree(true);

            if (tree.get_selected().length > 0) return

            if (stdoutDiv.data("stdout")) {
                stdoutDiv.trigger("click")
            } else if (errDiv.data("err")) {
                errDiv.trigger("click")
            }
        })
}

function updateAuxData(stdoutDiv, errDiv, stdout, err) {
    stdoutDiv.removeClass("aux-selected")
    errDiv.removeClass("aux-selected")

    stdoutDiv.data("stdout", stdout)
    errDiv.data("err", err)

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
}

function updateTest(updateBody) {
    $.ajax({
        type: "POST",
        url: "/update/test",
        data: JSON.stringify(updateBody),
        contentType: "application/json; charset=utf-8",
        success: function (data) {
            // alert(data);
        },
        error: function (errMsg) {
            alert("ERROR: " + errMsg.status);
        },
    });
}

function extractChildData(tree) {
    var childData = {}
    console.log("INTERNAL", tree._model.data)
    for (const [id, node] of Object.entries(tree._model.data)) {
        if (id == "#") continue
        childData[id] = (({ id, data, parent, text, type }) => ({ id, data, parent, text, type }))(node);
    }

    return childData
}