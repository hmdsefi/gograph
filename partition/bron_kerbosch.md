# gograph

## Clique Bron–Kerbosch Algorithm with Pivot + Degeneracy


#### Stepwise Visual (Table of R, P, X)
| Step | R (current clique) | P (candidates)      | X (processed) | Action / Clique Found         |
| ---- | ------------------ | ------------------- | ------------- | ----------------------------- |
| 1    | {}                 | {A,B,C,D,E,F,G,H,I} | {}            | Start recursion               |
| 2    | {A}                | {B,C,D,E,F,G,H,I}   | {}            | pivot B, expand C             |
| 3    | {A,C}              | {B,D}               | {}            | expand D                      |
| 4    | {A,C,D}            | {B}                 | {}            | expand B → max clique {A,B,C} |
| 5    | {E}                | {D,F}               | {}            | expand D → {E,D,F} max clique |
| 6    | {G}                | {H,I}               | {}            | expand H → {G,H,I} max clique |
