function NewNodeListener(id,defaultData={input:{},output:{}}) {
    var handlers = {
        id: id,
        onUpdate: function (data) {
            console.log(this.id + " (UPDATE)", data)
        },
        onRename: function (data) {
            console.log(this.id + " (RENAME)", data)
        },
        onDelete: function (data) {
            console.log(this.id + " (DELETE)", data)
        }
    }

   var tree = $(`#${id}`)
   
   tree.on('set_id.jstree', function (e, data) {
        if (data.node.data && data.node.data.isNew) {
            data.node.data = defaultData
            handlers.onUpdate(data, selectedId)
        } else if (data.old !== data.new) {
            handlers.onRename(data, selectedId)
        }
    });
    tree.on('delete_node.jstree', function (e, data) {
        handlers.onDelete(data, selectedId)
    });

    return handlers
}
$(function() {
    // replace jstree.set_id to properly update ids on renames
    $.jstree.core.prototype.set_id = function (obj, id) {
        obj = this.get_node(obj);
        if(!obj || obj.id === $.jstree.root) { return false; }
        var i, j, m = this._model.data, old = obj.id, oldLength = obj.id.length;
        id = id.toString();
        // update parents (replace current ID with new one in children and children_d)
        m[obj.parent].children[$.inArray(obj.id, m[obj.parent].children)] = id;

        var idMap = {[old]: id}
        for(i = 0, j = obj.children_d.length; i < j; i++) {
            var child = obj.children_d[i]
            if (idMap[child]) {
                continue
            }

            idMap[child] = child.replace(old, id)
        }

        for (const oldId of Object.keys(idMap)) {
            m[oldId].parent = idMap[m[oldId].parent] || m[oldId].parent
            m[oldId].parents = m[oldId].parents.map(function(parent) {
                return idMap[parent] || parent
            })
            m[oldId].children = m[oldId].children.map(function(child) {
                return idMap[child] || child
            })
            m[oldId].children_d = m[oldId].children_d.map(function(child) {
                return idMap[child] || child
            })
        }

        for(i = 0, j = obj.parents.length; i < j; i++) {
            m[obj.parents[i]].children_d = m[obj.parents[i]].children_d.map(function(child) {
                return idMap[child] || child
            })
        }

        i = $.inArray(obj.id, this._data.core.selected);
        if(i !== -1) { this._data.core.selected[i] = id; }

        for (const [oldId, newId] of Object.entries(idMap)) {
            i = this.get_node(oldId, true);
            if (i) {
                i.attr('id', newId); //.children('.jstree-anchor').attr('id', id + '_anchor').end().attr('aria-labelledby', id + '_anchor');
                if(this.element.attr('aria-activedescendant') === oldId) {
                    this.element.attr('aria-activedescendant', newId);
                }
            }

            Object.defineProperty(m, newId,
                Object.getOwnPropertyDescriptor(m, oldId));
            delete m[oldId]
            m[newId].id = newId;
            m[newId].li_attr.id = newId;
        }

        /**
         * triggered when a node id value is changed
         * @event
         * @name set_id.jstree
         * @param {Object} node
         * @param {String} old the old id
         */
        this.trigger('set_id',{ "node" : m[obj.id], "new" : obj.id, "old" : old });
        return true;
    }
})