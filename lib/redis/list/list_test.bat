@echo off
REM Assuming redis-cli -p 6389 is in the system path and Redis server is running on default port

echo Starting the test script...

REM Resetting the test environment by deleting old list if exists
REM echo Deleting existing list "mylist"...
REM redis-cli -p 6389 DEL mylist

echo Adding elements to the list with RPUSH...
redis-cli -p 6389 RPUSH mylist 1 2 3 444 "ccr" 666

echo Displaying all elements in the list:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Adding elements to the front of the list with LPUSH...
redis-cli -p 6389 LPUSH mylist "front1" "front2"

echo Updated list after LPUSH:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Removing and printing the first three elements with LPOP...
FOR /L %%A IN (1,1,3) DO (
    echo Popping element from the front:
    redis-cli -p 6389 LPOP mylist
)

echo Displaying remaining elements in the list after three LPOPs:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Removing and printing the last element with RPOP...
echo Popping element from the end:
redis-cli -p 6389 RPOP mylist

echo Displaying the list after RPOP:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Adding more elements to the list...
redis-cli -p 6389 RPUSH mylist 9 8 "ccr2"

echo Displaying the modified list:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Checking the first element with LINDEX:
echo First element:
redis-cli -p 6389 LINDEX mylist 0

echo Removing and printing the first three elements again...
FOR /L %%A IN (1,1,3) DO (
    echo Popping element:
    redis-cli -p 6389 LPOP mylist
)

echo Displaying the final state of the list:
redis-cli -p 6389 LRANGE mylist 0 -1

echo Test script completed.
