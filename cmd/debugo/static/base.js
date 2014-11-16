var NewConnection = function(obj, callback) {
	var self = {
		callback: callback
	}
	
	// Flag denoting if web socket connection is active
	self.ready = false;
	
	// We shall use the location of the window if location is not defined.
	if (self.location === undefined) {
		self.location = "ws://" + window.location.hostname + ":" + window.location.port + "/ws";
	}
	
	self.ws = new WebSocket(self.location);
	
	self.ws.onopen = function () {
		self.ready = true;	
	}

	self.ws.onmessage = function(evt){
		self.callback(obj, evt.data);
	}
	
	self.ws.onerror = function(evt){
		console.log("error", evt);	
	}
	
	self.send = function(data) {
		console.log("sending", data);
		self.ws.send(JSON.stringify(data));	
	}
	
	return self;
}; 

var File = Backbone.Model.extend({});

var FileList = Backbone.Collection.extend({
	model: File
});

var Command = Backbone.Model.extend({
	export: function() {
		var response = this.toJSON();
		if (this.get("Id") !== undefined){
			response.Id = this.get("Id");
		} else {
			response.Id = this.cid;
		}
		return response;
	}
});

var CommandList = Backbone.Collection.extend({
	model: Command
});

var CommandView = Backbone.View.extend({
	el: $("#cmd"),
	
	initialize: function(options) {
		this.commands = new CommandList();
		this.conn = options.conn;
	},
	
	events: {
		'keypress': 'handle'
	},
	
	handle: function(e){
		// return
		if (e.which === 13) {
			this.send(this.$el.val());
		}
	}, 
	
	send: function(text) {
		var command = new Command({'Command': text});
		this.commands.add(command);
		this.conn.send(command.export());
	}, 
	
	getOrCreate: function(data) {
		if (data.Id === undefined) return
		var command = this.commands.get(data.Id);
		if (command === undefined) {
			command = new Command(command);
		}
		return command;
	}
});

var App = Backbone.View.extend({
	el: document,
	events: {
		"keypress": "handle"
	},
	
	initialize: function(){
		this.conn = NewConnection(this, this.listen);
		this.cmdbar = new CommandView({conn: this.conn});
	},
	
	handle: function(e){
		// ctrl+space to activate command bar
		if (e.which === 32) {
			if (!e.ctrlKey) return;
			this.cmdbar.$el.focus();
		}
	},
	
	listen: function(that, data) {
		var parsed = {};
		try {
			parsed = JSON.parse(data);
		} catch (e) {
			console.log(d, e);
		}
		var command = that.cmdbar.getOrCreate(parsed);
		if (command === undefined) {
			console.log(parsed);
		}
		console.log(command);
	}
});
