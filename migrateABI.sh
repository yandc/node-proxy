#!/bin/bash
 
#地址
host='127.0.0.1'
#源端口
src_port=6379
src_db=2
#目标端口
dest_port=6379
dest_db=9
#计数器，统计迁移了多少KEY，末尾打印
cnt=0
 
 
#遍历获取所有KEY
for k in `redis-cli -n $src_db -h $host -p $src_port keys "contract_abi:*"`
do
        redis-cli -n $src_db -h $host -p $src_port move $k $dest_db
        let cnt++
done
#打印迁移的KEY
echo $cnt
