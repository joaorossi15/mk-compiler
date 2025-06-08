let map = fn(arr, f) {
	let iterator = fn(arr, acc) {
		if(len(arr) == 0) {
			acc
		} else {
			iterator(tail(arr), push(acc, f(first(arr))));
		} 
	};
	iterator(arr, []);
};

let a = [1, 2, 3, 4];
let t = fn(x) { x * 3 };
map(a, t);
