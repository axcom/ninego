### 能复制/引用到项目中使用的通用小功能、小工具：

- log
	项目实现可自行替换任意日志器的日志接口，主要解决不同项目使用不同log输出不一致的问题。

- filelog
	文件日志，可按日志大小，日期分割文件。
	
- class
	模拟实现java\delphi等语言的类，支持继承和多态。
	
- expr
	高精度的表达式运算解析器，指定公式运算结果。
	
- error
	实现了「错误码 + 错误信息 + 调用栈」三位一体的错误模型，解决了原生 error 无错误码、无栈信息的痛点。
	实现了 Try/Raise/ProtectRun 对panic错误的捕获转换。
	
- pinyin
	一组汉字首拼处理函数(也包括五笔首码)

- dataset
	将sql.Rows的record数据转换为DataSet数据集struct(相当于**"内存表"**)

- skit
	手搓&收集整理来的一些常用函数

- syscl
	常用数据结构方法合集List, Strings, StringList, Buffer, TreeNode... 