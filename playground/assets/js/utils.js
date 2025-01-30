function toId(parent, text) {
    return parent == "#" ? text : parent + "/" + text
}

customMenu = function(fileLabel) {
    return function (o, cb) {
        return {
            "create_folder": {
                "separator_before": false,
                "separator_after": false,
                "_disabled": o.type != "dir",
                "label": "New Folder",
                "action": function (data) {
                    var inst = $.jstree.reference(data.reference),
                        obj = inst.get_node(data.reference);
                    inst.create_node(obj, { type: "dir", data: { isNew: true } }, "last", function (new_node) {
                        try {
                            inst.edit(new_node, "New Folder", function (n, status, cancelled, last) {
                                if (status && !n.text.includes("/")) {
                                    inst.set_id(n, toId(n.parent, n.text))
                                } else {
                                    inst.delete_node(n);
                                }

                                console.log(n, status, cancelled, last)
                            });
                        } catch (ex) {
                            setTimeout(function () { inst.edit(new_node); }, 0);
                        }
                    });
                }
            },
            "create_file": {
                "separator_before": false,
                "separator_after": true,
                "_disabled": o.type != "dir",
                "label": "New " + fileLabel,
                "action": function (data) {
                    var inst = $.jstree.reference(data.reference),
                        obj = inst.get_node(data.reference);
                    inst.create_node(obj, { type: "file", data: { isNew: true } }, "last", function (new_node) {
                        try {
                            inst.edit(new_node, "New", function (n, status, cancelled) {
                                inst.set_id(n, toId(n.parent, n.text))
                                console.log(n, status, cancelled)
                            });
                        } catch (ex) {
                            setTimeout(function () {
                                inst.edit(new_node, "New", function (n, status, cancelled) {
                                    console.log(n, status, cancelled)
                                });
                            }, 0);
                        }
                    });
                }
            },
            "rename": {
                "separator_before": false,
                "separator_after": false,
                "_disabled": o.state.disabled,
                "label": "Rename",
                /*!
                "shortcut"			: 113,
                "shortcut_label"	: 'F2',
                "icon"				: "glyphicon glyphicon-leaf",
                */
                "action": function (data) {
                    var inst = $.jstree.reference(data.reference),
                        obj = inst.get_node(data.reference);

                    inst.edit(obj, null, function (n, status, cancelled, last) {
                        if (status && !n.text.includes("/")) {
                            inst.set_id(n, toId(n.parent, n.text))
                            console.log(inst._model.data)
                        } else {
                            // inst.delete_node(n);
                        }
                    })
                    // inst.edit(obj);
                }
            },
            "remove": {
                "separator_before": false,
                "icon": false,
                "separator_after": false,
                "_disabled": o.state.disabled,
                "label": "Delete",
                "action": function (data) {
                    var inst = $.jstree.reference(data.reference),
                        obj = inst.get_node(data.reference);
                    if (inst.is_selected(obj)) {
                        inst.delete_node(inst.get_selected());
                    }
                    else {
                        inst.delete_node(obj);
                    }
                }
            },
            "ccp": {
                "separator_before": true,
                "icon": false,
                "separator_after": false,
                "_disabled": o.state.disabled,
                "label": "Edit",
                "action": false,
                "submenu": {
                    "cut": {
                        "separator_before": false,
                        "separator_after": false,
                        "label": "Cut",
                        "action": function (data) {
                            var inst = $.jstree.reference(data.reference),
                                obj = inst.get_node(data.reference);
                            if (inst.is_selected(obj)) {
                                inst.cut(inst.get_top_selected());
                            }
                            else {
                                inst.cut(obj);
                            }
                        }
                    },
                    "copy": {
                        "separator_before": false,
                        "icon": false,
                        "separator_after": false,
                        "label": "Copy",
                        "action": function (data) {
                            var inst = $.jstree.reference(data.reference),
                                obj = inst.get_node(data.reference);
                            if (inst.is_selected(obj)) {
                                inst.copy(inst.get_top_selected());
                            }
                            else {
                                inst.copy(obj);
                            }
                        }
                    },
                    "paste": {
                        "separator_before": false,
                        "icon": false,
                        "_disabled": function (data) {
                            console.log("DATA", data)
                            console.log(data.reference)
                            console.log("REF", $.jstree.reference(data.reference))
                            // return false
                            return !$.jstree.reference(data.reference).can_paste();
                        },
                        "separator_after": false,
                        "label": "Paste",
                        "action": function (data) {
                            var inst = $.jstree.reference(data.reference),
                                obj = inst.get_node(data.reference);
                            inst.paste(obj);
                        }
                    }
                }
            }
        };
    }
}

function rootMenu(fileLabel) {
    fileLabel = fileLabel ? fileLabel : "File";
    return {
        "create_folder": {
            "separator_before": false,
            "separator_after": false,
            "_disabled": false,
            "label": "New Folder",
            "action": function (data) {
                var inst = $.jstree.reference(data.reference.prevObject[0]),
                    obj = inst.get_node("#");
                console.log(inst)
                inst.create_node(obj, { type: "dir", data: { isNew: true } }, "last", function (new_node) {
                    try {
                        inst.edit(new_node, "New Folder", function (n, status, cancelled, last) {
                            if (status && !n.text.includes("/")) {
                                inst.set_id(n, toId(n.parent, n.text))
                            } else {
                                inst.delete_node(n);
                            }

                            console.log(n, status, cancelled, last)
                        });
                    } catch (ex) {
                        setTimeout(function () { inst.edit(new_node); }, 0);
                    }
                });
            }
        },
        "create_file": {
            "separator_before": false,
            "separator_after": true,
            "_disabled": false,
            "label": "New " + fileLabel,
            "action": function (data) {
                var inst = $.jstree.reference(data.reference.prevObject[0]),
                    obj = inst.get_node("#");
                inst.create_node(obj, { type: "file", data: { isNew: true, input:{},expected:{} } }, "last", function (new_node) {
                    try {
                        inst.edit(new_node, "New", function (n, status, cancelled) {
                            inst.set_id(n, toId(n.parent, n.text))
                            console.log("CALLBACK")
                            console.log(n, status, cancelled)
                        });
                    } catch (ex) {
                        setTimeout(function () {
                            inst.edit(new_node, "New", function (n, status, cancelled) {
                                console.log("EX CALLBACK")
                                console.log(n, status, cancelled)
                            });
                        }, 0);
                    }
                });
            }
        }
    }
}

function sortFunc (a, b) {
    a1 = this.get_node(a);
    b1 = this.get_node(b);
    if (a1.type == b1.type) {
      return a1.text > b1.text ? 1 : -1;
    } else {
      return a1.type > b1.type ? 1 : -1;
    }
  }
