2048-bonacci
2048 is a tiny video game that was super hype some years ago. You had to fuse two equal numbers to create its double. So 1 and 1 became 2, 2 and 2 became 4, and so on.

Fibonacci was a famous mathematician who made the super-hype Fibonacci sequence you may have heard about: U(n+1) = U(n) + U(n-1). The first number of the sequence are 1, 1, 2, 3, 5, 8, 13, …

Let’s combine these two hype things to create a super-super-hype game!

2048-bonacci plays on a 4x4 square. Each square is either empty or contains a number of the Fibonacci sequence.

You are given an initial board situation (a 2D array of integers) and a pushing direction (up, left, down, or right). Then, you must compute the board contents after the push and return an updated 2D array of integers.

The value of integers in the array is guaranteed to be less than 2^16 and are all Fibonacci numbers. The value 0 means the square is empty.

Rule 1
Numbers move as far as possible in the pushing direction.

-------------                       -------------
| 2|  |  |  |                       |  |  |  |  |
-------------                       -------------
|  |  |13|  |                       |  |  |  |  |
-------------  => push downward =>  -------------
|  |  |  |  |                       | 2|  |  |  |
-------------                       -------------
| 5|  |  |  |                       | 5|  |13|  |
-------------                       -------------

Rule 2
When two consecutive numbers in the Fibonacci sequence are pushed one on another, they fuse into the next number.

-------------                       -------------
|  |  | 1| 2|                       |  |  |  | 3|
-------------                       -------------
| 1|  | 1|  |                       |  |  |  | 2|
-------------  => push rightward => -------------
|  | 8| 5|  |                       |  |  |  |13|
-------------                       -------------
|  | 5| 8|  |                       |  |  |  |13|
-------------                       -------------
Rule 3
Fusing orders are resolved in the backward direction of the push.

A fused number can not be fused once again in the same turn.

-------------                 -------------                 -------------
|  | 1| 2| 3|                 |  |  | 1| 5|                 |  |  | 1| 5|
-------------                 -------------                 -------------
|  | 3| 2| 1|                 |  |  | 3| 3|                 |  |  | 3| 3|
------------- => rightward => ------------- => rightward => -------------
|  |  |  |  |                 |  |  |  |  |                 |  |  |  |  |
-------------                 -------------                 -------------
|  | 5| 3| 5|                 |  |  | 5| 8|                 |  |  |  |13|
-------------                 -------------                 -------------
Rule 4
Numbers can move to a square that a fusing has just emptied.

-------------                     -------------
| 1|  |  |  |                     | 2|  |  |  |
-------------                     -------------
| 1|  |  |  |                     | 2|  |  |  |
-------------  => push upward =>  -------------
| 1|  |  |  |                     |  |  |  |  |
-------------                     -------------
| 1|  |  |  |                     |  |  |  |  |
-------------                     -------------
Let’s hype!