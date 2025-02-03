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
  
  for (const [id, node] of Object.entries(tree._model.data)) {
    if (id == "#") continue
    childData[id] = (({ id, data, parent, text, type }) => ({ id, data, parent, text, type }))(node);
  }

  return childData
}

function NewChildTree(conf) {
  var childId = "#" + conf.id;
  var parentId = "#" + conf.parent;
  var events = NewNodeListener(conf.id, "")

  events.onUpdate = function(data, parent) {
    var pnode = $(parentId).jstree(true).get_node(parent) 
    var picked = (({ id, data, parent, text, type }) => ({ id, data, parent, text, type }))(data.node);

    pnode.data[conf.src][data.node.id] = picked

    updateTest({
      op: data.node.type == "dir" ? "PUSH_DIR" : "PUSH",
      id: data.node.id,
      parentId: parent,
      target: conf.src,
      content: data.node.data,
    });
  }

  events.onRename = function(data, parent) {
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

  events.onDelete = function(data, parent) {
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

  $(childId).jstree({
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
    plugins: ["contextmenu", "types", "sort", "unique"],
    contextmenu: {
      items: customMenu("File"),
    },
  });

  var prompt = $(childId).prev().children(":first").data("prompt")
  $(childId)
    .on("refresh.jstree", function (e, data) {
      if (Object.keys(data.instance._model.data).length == 1) {
        prev = $(this).prev();
        inner = prev.children(":first");
        if (selectedId == "") {
          inner.text(prompt);
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
        model.id = data.node.id;

        var updateContent = debounce(function() {
          var value = this.model.getValue()
          this.data.node.data = value
          // this.parent.data[this.src][this.model.id].data = value;
          this.events.onUpdate(this.data, this.parent.id)
        }.bind({
          data: data,
          parent: $(parentId).jstree(true).get_node(selectedId),
          src: conf.src,
          model: model,
          events: events,
        }));

        model.onDidChangeContent(() => {
          updateContent()
        });

        conf.editor.setModel(model);
        conf.editor.updateOptions({ readOnly: false });
      }
    })
    .on("contextmenu", function (e) {
      if (selectedId == "" || e.target.id != conf.id) {
        return;
      }
      e.preventDefault();

      tree = $(this).jstree(true);
      tree.deselect_all(true);
      obj = tree.get_node("#");
      tree._show_contextmenu(obj, e.pageX, e.pageY, rootMenu());
    });

  $(parentId).on("select_change", function (e, data) {
    var tree = $(childId).jstree(true);
    tree.settings.core.data = data.node.type == "file" ? Object.values(data.node.data[conf.src]) : null;
    tree.refresh(true, true);

    clearEditor(conf.editor);
  });
}
