let name = "placeholder";

let greeter = fn(greet) {
	fn(name) {
		return greet + " " + name + "!"; 	
	}
}

let hi = greeter("hi");
hi("john");
