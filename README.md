# logcrane
a log recording component by golang

日志分为两种：
1.玩家日志（PlayerBaseLog）：针对记录玩家的个人行为
2.服务器日志（ServerBaseLog）：针对记录服务器的事件
所有日志都是继承自这两者之一

日志定义方案：
在任意包定义每个日志的结构，对应sql结构写在tag里面
tag：
type：字段对应日志表中列的类型；
length：字段对应日志表中列的长度，一般varchar类定义，如果没定义默认255。其他类型都不用定义，使用mysql的默认长度；
explain：字段的注释；
name：字段对应日志表中列名。如果没有指定，则默认是该字段名小写；

每种日志都有自己的一个channel，在新请求一个日志时，就会启动一个自己的Goroutine，
这个Goroutine会一直从自己的channel中取出日志，取出的数量根据定义日志的插入类型是single还是batch决定

调用方法：
crane.Start(serverId) // 启动日志系统
crane.Instance().Execute(Logger) // 执行日志记录
