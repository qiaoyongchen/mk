# mk

运行:
go run main.go

```ocaml
let map = fn(arr, f) {
    let iter = fn(arr, acc) {
        if (len(arr) == 0) {
            return acc;
        } else { 
            return iter(rest(arr), push(acc, f(first(arr)))); 
        }
    };
    return iter(arr, []);
};
   
map([1,2,3,4,5], fn(x) { return x * 2; });

let a = 1 + 2 + 3 * 4 * (5 + 6);

let b = [1,2,3,4,fn(x) {return x;}, 5,6,7];
b[0]
b[4](5)

let c = {11:"11", 22:"22", 11+22:"33", 44:[1,2,3,4,5]};
c[44]  
```
