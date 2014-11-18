var Command = Backbone.Model.extend({	
	initialize: function(){
		this.history = [];
	},
	
	add: function(command, args) {
		this.history.unshift({
			Command: command,
			Args: args
		});
	},
	
	export: function() {
		var response = this.history[0];
		if (this.id !== undefined){
			response.Id = this.id;
		} else {
			response.Id = this.cid;
		}
		return response;
	}
});

var CommandList = Backbone.Collection.extend({
	model: Command, 
	
	register: function(command, args) {
		var comm = new Command();
		comm.add(command, args);
		this.add(comm);
		return comm;
	},

	getOrCreate: function(args) {
		var id = args.Id;
		if (id === undefined) {
			return this.register(args.Command, args.Args);
		}
		var comm = this.get(id);
		if (comm !== undefined) {
			comm.add(args.Command, args.Args);
			return comm;
		} 
		var comm = this.register(args.Command, args.Args);
		comm.id = id;
		return comm;
	}
});


var File = Backbone.Model.extend({
	getContent: function(callback) {
		if (this.content !== undefined){
			return;
		}
		var comm = Connection.commands.register("getfile", [this.id]);
		comm.callback = this.setContent.bind(this);
		this.callback = callback;
		Connection.send(comm.export());
	},

	setContent: function(comm) {
		this.content = comm.export().Args[0];
		this.callback();
	}
});

var FileList = Backbone.Collection.extend({
	model: File, 
	
	populate: function(data) {		
		var that = this;
		_.each(data, function(file) {
			var f = new File({"id": file});
			that.add(f);
		});
	}
});


var Connection = {
	ready: false,
	
	init: function(){
		var that = this;
		this.location = "ws://" + window.location.hostname + ":" + window.location.port + "/ws";
		this.ws = new WebSocket(this.location);
		this.ws.onopen = function() {
			that.ready = true;
		};
		this.ws.onmessage = function(evt){
			that.handle(evt.data);
		};
		this.ws.onerror = function(evt){
			console.log("error", evt);	
		};

		this.commands = new CommandList();
		
		this.callbacks = {};
	}, 
	
	send: function(data) {
		this.ws.send(JSON.stringify(data));
	}, 

	// register sets a callback method to be executed onmessage event of the WebSocket.
	register: function(obj, method, name) {
		this.callbacks[name] = method.bind(obj);
	},
	
	handle: function(data) {
		var parsed = {};
		try {
			parsed = JSON.parse(data);
		} catch (e) {
			console.log(d, e);
			return;
		}
		var command = this.commands.getOrCreate(parsed);
		if (command === undefined) {
			console.log(parsed);
		}
		var fn;
		if(command.callback !== undefined){
			fn = command.callback;
		} else {
			 fn = this.callbacks[command.export().Command];
		}
		if (fn === undefined) return;
		
		fn(command);
	},
}


var FileContentView = Backbone.View.extend({
	tagName: "table",
	
	events: {
		"click .lineno": "setbreakpoint"
	},
	
	template: _.template($("#tmpl-file").html()),
	
	initialize: function(){
		this.model.contentview = this;
	},

	render: function() {
		var lines = this.model.content.split("\n");
		for (var i=0; i<=lines.length; i++) {
			this.$el.append(this.template({line: lines[i], number: i+1}));
		}
		return this;
	},
	
	setbreakpoint: function(e) {
		var line = e.target.innerHTML;
		if (line == "") {
			return;
		}
		var comm = Connection.commands.register("b", [this.model.id + ":" + line]);
		comm.callback = this.markbreakpoint.bind(this);
		Connection.send(comm.export());
	},
	
	markbreakpoint: function(comm) {
		var args = comm.export().Args;
		// The response is [filename, lineno]
		if (this.model.id !== args[0]) {
			console.log("Files are not the same.")
			return
		}
		var lineno = args[1];
		var el = this.$el.find(".line" + lineno);
		el.addClass("bp");
	},
	
	highlight: function(line) {
		this.$el.find(".highlight").removeClass("highlight");
		var el = this.$el.find(".textlineno" + line);
		el.addClass("highlight");
	}
});

var FileNameView = Backbone.View.extend({
	tagName: "p",
	
	className: "file",
	
	$container: $("#content"),
	
	events: {
		"click": "detailed"
	},
	
	initialize: function(){
		this.model.nameview = this;
	},
	
	render: function(){
		this.$el.html(this.model.id);
	},
	
	detailed: function() {
		if (this.model.content !== undefined) {
			if (this.view === undefined) {
				this.view = new FileContentView({model: this.model});		
			}
			this.$container.html(this.view.render().el);
		} else {
			this.model.getContent(this.detailed.bind(this));
		}
	},
	
	highlight: function() {
		this.$el.find(".highlight").removeClass("highlight");
		this.$el.addClass("highlight");
	}
});

var FileListView = Backbone.View.extend({	
	initialize: function(){
		this.render();
	},
	
	render: function(){
		var that = this;
		that.model.each(function(file){
			that.addOne(file);
		});
		return that;
	},
	
	addOne: function(file) {
		var view = new FileNameView({model: file});
		view.render();
		this.$el.append(view.el);	
	}
});


var CommandView = Backbone.View.extend({
	el: $("#cmd"),
	
	events: {
		'keypress': 'handle'
	},
	
	handle: function(e){
		// return
		if (e.which === 13) {
			this.send(this.$el.val());
			this.$el.val("");
		}
	}, 
	
	send: function(text) {
		var args = text.split(" ");
		var command = Connection.commands.register(args[0], args.slice(1));
		Connection.send(command.export());
	}
});

var App = Backbone.View.extend({
	el: document,
	events: {
		"keypress": "handle"
	},
	
	initialize: function(){
		Connection.init();
		Connection.register(this, this.popFiles, "popFiles");
		Connection.register(this, this.paused, "paused");
		Connection.register(this, this.exited, "exited");
		Connection.register(this, this.error, "error");
		this.cmdbar = new CommandView();
	},
	
	handle: function(e){
		// ctrl+space to activate command bar
		if (e.which === 32) {
			if (!e.ctrlKey) return;
			this.cmdbar.$el.focus();
		}
	},
	
	// popFiles populates the filesytem sidebar.
	popFiles: function(cmd) {
		this.files = new FileList();
		this.files.populate(cmd.export().Args);
		this.fileView = new FileListView({model: this.files});
		$("#filesystem").html(this.fileView.el);
	},
	
	error: function(args) {
		_.each(args, function(arg) {console.log(arg);});
	}, 
	
	paused: function(cmd) {
		var obj = cmd.export();
		var file = this.fileView.model.get(obj.Args[0]);
		var nameView = file.nameview;
		var contentView = file.contentview;
		$("#content").html(contentView.el);
		contentView.highlight(obj.Args[1]);
		nameView.highlight();
	}, 
	
	exited: function(cmd) {
		alert("Hurrah! The command completed.");
	}
});
