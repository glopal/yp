$(document).on("update_node", function (e, data) {
    console.log("update_node event: ", data)
});

function NewNodeListener(id) {
    handlers = {
        id: id,
        onUpdate: function (data) {
            console.log(this.id + " (UPDATE)", data)
        },
        onRename: function (data) {
            console.log(this.id + " (RENAME)", data)
        }
    }

    $(`#${id}`).on('rename_node.jstree', function (e, data) {
        if (data.node.data && data.node.data.isNew) {
            data.node.data = {}
            handlers.onUpdate(data)
        } else if (data.old !== data.text) {
            handlers.onRename(data)
        }
    });

    return handlers
}