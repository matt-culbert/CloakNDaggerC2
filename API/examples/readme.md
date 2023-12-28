This holds examples for running different commands.

If you want to automate something or write your own interaction with the API for whatever reason, these are examples of how you can do it.

Set.go will set/hset a block of implant info in the redis DB. I loaded the DB with test1,2,3 for queries

Hget.go will hget the results for a hashed implant ID, all it needs is the implant UUID. In the examples, I used test1,2,3 etc