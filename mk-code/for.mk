let a = fn(b) {
	if(b == 300000) {
		return b;
	}
	if(b < 3000000) {
		a(b+1);
	}
}

a(1);
