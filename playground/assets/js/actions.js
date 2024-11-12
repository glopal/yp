function sync(inst) {
    // tree = $("#tests").jstree(true)

    // inst.get_json("#", {
    //     flat: true
    // }).forEach

    // $(tree.get_json("#", {
    //     flat: true
    // })).each(function () {
    //     // Get the level of the node
    //     // var level = tree.get_node(this.id)
    //     console.log("NODE", tree.get_node(this.id))
    //     // var node;
    //     // if (level == 3) {
    //     //     // node = ... apply desired css to the node here if it has children.
    //     // }
    // });
    treeData = []
    inst.get_json("#", {
        flat: true
    }).forEach((node) => {
        if (node.id == "#") {
            return
        }
        treeData.push({
            id: node.id,
            type: node.type,
            data: node.data,
        })
    });
    console.log("TREE", treeData)
    $.ajax({
        type: "POST",
        url: '/sync',
        dataType: 'json',
        async: false,
        data: JSON.stringify(treeData),
        success: function () {

        }
    })
}