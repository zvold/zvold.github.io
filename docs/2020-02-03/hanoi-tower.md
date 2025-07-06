---
layout: math
---

# Hanoi Tower

There is a nice recursive solution to the [Tower of Hanoi](https://en.wikipedia.org/wiki/Tower_of_Hanoi) problem.

```
 1                         1
 2           --->          2
_3___.___._       _.___.___3_
 i   k   j         i   k   j
```

Imagine there is a function $$F(N, i, j)$$ that can properly move a tower of height $$N$$ (that is, disk $$1$$on top of $$2$$, on top of ... $$N$$) from position $$i$$ to $$j$$, using the “free” position $$k$$. The function $$F$$ operates on the full “state” of the pins and outputs a new “state”.

Note that the position $$k$$ doesn’t have to be truly free — since it can only have disks of size more than $$N$$ — it’s *effectively* free. This is because we’re assuming that the “state” is always correct (the disks are placed according to the rules).

Then, implementation of $$F(N+1,i,j)$$ is:

1. Apply $$F(N,i,k)$$ — move the stack of height $$N$$ from $$i$$ to $$k$$. This leaves the $$(N+1)$$ disk at $$i$$.
2. Move the disk $$(N+1)$$ from $$i$$ to $$j$$ (because after step 1, position $$j$$ is *effectively* empty).
3. Apply $$F(N, k, j)$$ — move the stack of height $$N$$ from $$k$$ on top of disk $$(N+1)$$ at $$j$$.

Recursion terminates at $$F(1, i, j)$$, which is trivial to implement — the disk of size $$1$$ (the smallest) can always be moved anywhere.

---

So the sequence of calls for the picture above becomes:

Call stack:

1. \$$F(3,i,j)$$
    1. \$$F(2,i,k)$$
        1. \$$F(1,i,j)$$
        2. move from $$i$$ to $$k$$
        3. \$$F(1,j,k)$$
    2. move from $$i$$ to $$j$$
    3. \$$F(2,k,j)$$
        1. \$$F(1, k, i)$$
        2. move from $$k$$ to $$j$$
        3. \$$F(1, i, j)$$

State:
1. `[1 2 3]   [ ]      [ ]`
    1. no changes
        1. `[2 3]     [ ]      [1]`
        2. `[3]       [2]      [1]`
        3. `[3]     [1 2]      [ ]`
    2. `[ ]     [1 2]      [3]`
    3. no changes
        1. `[1]       [2]      [3]`
        2. `[1]       [ ]    [2 3]`
        3. `[ ]       [ ]  [1 2 3]`

In this sequence, only some steps do the actual moves: $$F(1,\_,\_)$$ and “move from _ to _”. The former can always be performed (because it’s the disk of size $$1$$). But it’s not obvious why the latter can always be performed.

Intuitively:

- (1.1.2) can be performed, because it moves the disk of size $$2$$ to $$k$$. We know that the only disk can prevent it (size $$1$$) is not at position $$k$$, because it was moved to $$j$$ in previous step (1.1.1).
- (1.3.2) can be performed for the same reason.
- (1.2) can be performed, because it moves the disk of size $$3$$ to $$j$$. The disks that can prevent it (sizes $$2$$ and $$1$$) are not at position $$j$$, because they both were just moved to $$k$$ in step (1.1).
