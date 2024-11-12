customMenu = function (o, cb) {
    return {
        "create_folder": {
            "separator_before": false,
            "separator_after": false,
            "_disabled": o.type != "dir",
            "label": "New Folder",
            "action": function (data) {
                console.log(data.reference)
                var inst = $.jstree.reference(data.reference),
                    obj = inst.get_node(data.reference);
                console.log(inst)
                inst.create_node(obj, { type: "dir" }, "last", function (new_node) {
                    try {
                        inst.edit(new_node, "New Folder", function (n, status, cancelled, last) {
                            if (status && !n.text.includes("/")) {
                                inst.set_id(n, n.parent + "/" + n.text)
                                $.ajax
                                    ({
                                        type: "POST",
                                        url: '/create',
                                        dataType: 'json',
                                        async: false,
                                        //json object to sent to the authentication url
                                        data: JSON.stringify({
                                            id: n.id,
                                            type: n.type,
                                            parent: n.parent,
                                            text: n.text
                                        }),
                                        success: function () {

                                        }
                                    })
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
            "label": "New Test",
            "action": function (data) {
                var inst = $.jstree.reference(data.reference),
                    obj = inst.get_node(data.reference);
                inst.create_node(obj, { type: "file" }, "last", function (new_node) {
                    try {
                        inst.edit(new_node, "New", function (n, status, cancelled) {
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
        },
        "rename": {
            "separator_before": false,
            "separator_after": false,
            "_disabled": false, //(this.check("rename_node", data.reference, this.get_parent(data.reference), "")),
            "label": "Rename",
            /*!
            "shortcut"			: 113,
            "shortcut_label"	: 'F2',
            "icon"				: "glyphicon glyphicon-leaf",
            */
            "action": function (data) {
                var inst = $.jstree.reference(data.reference),
                    obj = inst.get_node(data.reference);
                console.log($.jstree.reference(data.reference).get_json('#', { flat: true }))

                inst.edit(obj, null, function (n, status, cancelled, last) {
                    console.log(Object.keys(inst))
                    if (status && !n.text.includes("/")) {
                        inst.set_id(n, n.parent + "/" + n.text)
                        sync(inst)

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
            "_disabled": false, //(this.check("delete_node", data.reference, this.get_parent(data.reference), "")),
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

createFolderAction = {
    "separator_before": false,
    "separator_after": false,
    "_disabled": false, //(this.check("create_node", data.reference, {}, "last")),
    "label": "New Folder",
    "action": function (data) {
        var inst = $.jstree.reference(data.reference),
            obj = inst.get_node(data.reference);
        inst.create_node(obj, { type: "dir" }, "last", function (new_node) {
            try {
                inst.edit(new_node);
            } catch (ex) {
                setTimeout(function () { inst.edit(new_node); }, 0);
            }
        });
    }
}

createFileAction = {
    "separator_before": false,
    "separator_after": true,
    "_disabled": false, //(this.check("create_node", data.reference, {}, "last")),
    "label": "New Test",
    "action": function (data) {
        var inst = $.jstree.reference(data.reference),
            obj = inst.get_node(data.reference);
        inst.create_node(obj, { type: "file" }, "last", function (new_node) {
            try {
                inst.edit(new_node);
            } catch (ex) {
                setTimeout(function () { inst.edit(new_node); }, 0);
            }
        });
    }
}
baseActions = {
    "create_folder": {
        "separator_before": false,
        "separator_after": false,
        "_disabled": false, //(this.check("create_node", data.reference, {}, "last")),
        "label": "New Folder",
        "action": function (data) {
            var inst = $.jstree.reference(data.reference),
                obj = inst.get_node(data.reference);
            inst.create_node(obj, { type: "dir" }, "last", function (new_node) {
                try {
                    inst.edit(new_node);
                } catch (ex) {
                    setTimeout(function () { inst.edit(new_node); }, 0);
                }
            });
        }
    },
    "create_file": {
        "separator_before": false,
        "separator_after": true,
        "_disabled": false, //(this.check("create_node", data.reference, {}, "last")),
        "label": "New Test",
        "action": function (data) {
            var inst = $.jstree.reference(data.reference),
                obj = inst.get_node(data.reference);
            inst.create_node(obj, { type: "file" }, "last", function (new_node) {
                try {
                    inst.edit(new_node);
                } catch (ex) {
                    setTimeout(function () { inst.edit(new_node); }, 0);
                }
            });
        }
    },
    "rename": {
        "separator_before": false,
        "separator_after": false,
        "_disabled": false, //(this.check("rename_node", data.reference, this.get_parent(data.reference), "")),
        "label": "Rename",
        /*!
        "shortcut"			: 113,
        "shortcut_label"	: 'F2',
        "icon"				: "glyphicon glyphicon-leaf",
        */
        "action": function (data) {
            var inst = $.jstree.reference(data.reference),
                obj = inst.get_node(data.reference);
            inst.edit(obj);
        }
    },
    "remove": {
        "separator_before": false,
        "icon": false,
        "separator_after": false,
        "_disabled": false, //(this.check("delete_node", data.reference, this.get_parent(data.reference), "")),
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