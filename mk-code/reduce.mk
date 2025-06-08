let reduce = fn(arr, total, f) {
	let iterator = fn(arr, res) {
		if(len(arr) == 0) {
			res
		} else {
			iterator(tail(arr), f(res, first(arr)));	
		}
	};
	iterator(arr, total);
};

let sum = fn(arr) {
	reduce(arr, 0, fn(total, e) { total + e });
};

let a = [1, 2, 3];
sum(a);
