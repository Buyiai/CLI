# 1、概述
CLI（Command Line Interface）实用程序是Linux下应用开发的基础。正确的编写命令行程序让应用与操作系统融为一体，通过shell或script使得应用获得最大的灵活性与开发效率。Linux提供了cat、ls、copy等命令与操作系统交互；go语言提供一组实用程序完成从编码、编译、库管理、产品发布全过程支持；容器服务如docker、k8s提供了大量实用程序支撑云服务的开发、部署、监控、访问等管理任务；git、npm等都是大家比较熟悉的工具。尽管操作系统与应用系统服务可视化、图形化，但在开发领域，CLI在编程、调试、运维、管理中提供了图形化程序不可替代的灵活性与效率。

# 2、基础知识
## selpg 程序逻辑
selpg 是从文本输入选择页范围的实用程序。该输入可以来自作为最后一个命令行参数指定的文件，在没有给出文件名参数时也可以来自标准输入。

selpg 首先处理所有的命令行参数。在扫描了所有的选项参数后，如果 selpg 发现还有一个参数，则它会接受该参数为输入文件的名称并尝试打开它以进行读取。如果没有其它参数，则 selpg 假定输入来自标准输入。

### 参数处理
**“-sNumber”和“-eNumber”强制选项：**

selpg 要求用户用两个命令行参数“-sNumber”（例如，“-s10”表示从第 10 页开始）和“-eNumber”（例如，“-e20”表示在第 20 页结束）指定要抽取的页面范围的起始页和结束页。selpg 对所给的页号进行合理性检查；换句话说，它会检查两个数字是否为有效的正整数以及结束页是否不小于起始页。这两个选项，“-sNumber”和“-eNumber”是强制性的，而且必须是命令行上在命令名 selpg 之后的头两个参数：
```
$ selpg -s10 -e20 ...
```
（... 是命令的余下部分，下面对它们做了描述）。

**“-lNumber”和“-f”可选选项：**

selpg 可以处理两种输入文本：

*类型 1：* 该类文本的页行数固定。这是缺省类型，因此不必给出选项进行说明。也就是说，如果既没有给出“-lNumber”也没有给出“-f”选项，则 selpg 会理解为页有固定的长度（每页 72 行）。

选择 72 作为缺省值是因为在行打印机上这是很常见的页长度。这样做的意图是将最常见的命令用法作为缺省值，这样用户就不必输入多余的选项。该缺省值可以用“-lNumber”选项覆盖，如下所示：
```
$ selpg -s10 -e20 -l66 ...
```
这表明页有固定长度，每页为 66 行。

*类型 2：* 该类型文本的页由 ASCII 换页字符（十进制数值为 12，在 C 中用“\f”表示）定界。该格式与“每页行数固定”格式相比的好处在于，当每页的行数有很大不同而且文件有很多页时，该格式可以节省磁盘空间。在含有文本的行后面，类型 2 的页只需要一个字符 ― 换页 ― 就可以表示该页的结束。打印机会识别换页符并自动根据在新的页开始新行所需的行数移动打印头。

类型 2 格式由“-f”选项表示，如下所示：
```
$ selpg -s10 -e20 -f ...
```
该命令告诉 selpg 在输入中寻找换页符，并将其作为页定界符处理。

注：“-lNumber”和“-f”选项是互斥的。

**“-dDestination”可选选项：**

selpg 还允许用户使用“-dDestination”选项将选定的页直接发送至打印机。这里，“Destination”应该是 lp 命令“-d”选项可接受的打印目的地名称。该目的地应该存在 ― selpg 不检查这一点。在运行了带“-d”选项的 selpg 命令后，若要验证该选项是否已生效，请运行命令“lpstat -t”。该命令应该显示添加到“Destination”打印队列的一项打印作业。如果当前有打印机连接至该目的地并且是启用的，则打印机应打印该输出。这一特性是用 popen() 系统调用实现的，该系统调用允许一个进程打开到另一个进程的管道，将管道用于输出或输入。在下面的示例中，打开到命令
```
$ lp -dDestination
```
的管道以便输出，并写至该管道而不是标准输出：
```
selpg -s10 -e20 -dlp1
```
该命令将选定的页作为打印作业发送至 lp1 打印目的地。可以看到类似“request id is lp1-6”的消息。该消息来自 lp 命令；它显示打印作业标识。如果在运行 selpg 命令之后立即运行命令 lpstat -t | grep lp1 ，应该看见 lp1 队列中的作业。如果在运行 lpstat 命令前耽搁了一些时间，那么可能看不到该作业，因为它一旦被打印就从队列中消失了。

### 输入处理
一旦处理了所有的命令行参数，就使用这些指定的选项以及输入、输出源和目标来开始输入的实际处理。

selpg 通过以下方法记住当前页号：如果输入是每页行数固定的，则 selpg 统计新行数，直到达到页长度后增加页计数器。如果输入是换页定界的，则 selpg 改为统计换页符。这两种情况下，只要页计数器的值在起始页和结束页之间这一条件保持为真，selpg 就会输出文本（逐行或逐字）。当那个条件为假（也就是说，页计数器的值小于起始页或大于结束页）时，则 selpg 不再写任何输出。

# 3、开发实践
+ 引用到的包如下：
```go
import (
	"bufio" // bufio 用来帮助处理 I/O 缓存
	"fmt"
	"io"
	"os"
	"os/exec"

	flag "github.com/spf13/pflag"
)
```
+ 定义保存参数数据的结构体 selpgArgs 如下：
```go
type selpgArgs struct {
	startPage int // 开始页
	endPage   int // 结束页

	inFilename string // 输入文件名
	printDest  string // 输出文件名

	pageLen  int    // 每页的行数，默认为72
	pageType string // 'l'按行打印，'f'按换页符打印，默认按行
}
```
+ 声明用来保存程序名的全局变量，用于显示错误信息。
```go
var progname string // 保存名称（命令就是通过该名称被调用）的全局变量，作为在错误消息中显示之用
```
+ main 函数首先声明一个名为 sa 的 selpgArgs，然后使用 `os.Args`读取程序输入的所有参数，初始化 selpgArgs 里的各个参数，接着调用 processArgs 函数和 processInput 函数。具体代码如下：
```go
func main() {
	sa := selpgArgs{}
	progname = os.Args[0]

	processArgs(&sa) // 处理参数
	processInput(sa) // 处理输入输出
}
```
+ 函数 processArgs 主要是分析用户输入的命令，进行错误检查，判断每个参数的格式是否正确、参数个数是否正确，并将各种信息存储在 sa 中。用 pflag 绑定 sa 的各个参数，命令行中的信息就会自动存入 sa。参考：[Golang 之使用 Flag和 Pflag](https://o-my-chenjian.com/2017/09/20/Using-Flag-And-Pflag-With-Golang/)

首先将 flag 绑定到 sa 的各个参数上：
```go
	flag.IntVarP(&sa.startPage, "start", "s", -1, "start page(>1)")
	flag.IntVarP(&sa.endPage, "end", "e", -1, "end page(>=start_page)")
	flag.IntVarP(&sa.pageLen, "len", "l", 10, "page len")
	flag.StringVarP(&sa.printDest, "dest", "d", "", "print dest")
	flag.StringVarP(&sa.pageType, "type", "f", "l", "'l' for lines-delimited, 'f' for form-feed-delimited. default is 'l'")
	flag.Lookup("type").NoOptDefVal = "f"
```
第一个参数为变量，第二个参数为命令行参数名，第三个参数为该参数的简写，第四个参数为该参数没有在命令行出现时的默认值，第五个参数为帮助信息。

接着调用 `flag.Parse()` 解析命令行参数到定义的 flag，然后检查各个参数的合法性。

参数个数不够：
```go
	if len(os.Args) < 3 { // 参数个数不够（至少为 progname -s start_page -e end_page）
		fmt.Fprintf(os.Stderr, "\n%s: not enough arguments\n", progname)
		flag.Usage()
		os.Exit(1)
	}
```
处理第一个参数：
```go
	if os.Args[1] != "-s" {
		fmt.Fprintf(os.Stderr, "\n%s: 1st arg should be -s start_page\n", progname)
		flag.Usage()
		os.Exit(2)
	}

	intMax := 1<<32 - 1

	if sa.startPage < 1 || sa.startPage > intMax {
		fmt.Fprintf(os.Stderr, "\n%s: invalid start page %s\n", progname, os.Args[2])
		flag.Usage()
		os.Exit(3)
	}
```
处理第二个参数：
```go
	if os.Args[3] != "-e" {
		fmt.Fprintf(os.Stderr, "\n%s: 2nd arg should be -e end_page\n", progname)
		flag.Usage()
		os.Exit(4)
	}

	if sa.endPage < 1 || sa.endPage > intMax || sa.endPage < sa.startPage {
		fmt.Fprintf(os.Stderr, "\n%s: invalid end page %s\n", progname, sa.endPage)
		flag.Usage()
		os.Exit(5)
	}
```
处理每页行数：
```go
	if sa.pageLen < 1 || sa.pageLen > (intMax-1) {
		fmt.Fprintf(os.Stderr, "\n%s: invalid page length %s\n", progname, sa.pageLen)
		flag.Usage()
		os.Exit(5)
	}
```
检查输入文件：
```go
	if len(flag.Args()) == 1 {
		_, inFileErr := os.Stat(flag.Args()[0])
		// 检查文件是否存在
		if inFileErr != nil && os.IsNotExist(inFileErr) {
			fmt.Fprintf(os.Stderr, "\n%s: input file \"%s\" does not exist\n",
				progname, flag.Args()[0])
			os.Exit(6)
		}
		sa.inFilename = flag.Args()[0]
	}
```
+ 函数 processInput 首先设置输入源，即选择从哪里进行读取，然后设置输出源，即选择打印到哪里，接下来进行打印。

读取输入文件，若缺省，则通过标准输入（键盘或重定向）读取输入流。
```go
	var fin *os.File
	if len(sa.inFilename) == 0 {
		fin = os.Stdin
	} else {
		var inputError error
		fin, inputError = os.Open(sa.inFilename)
		if inputError != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open input file \"%s\"\n",
				progname, sa.inFilename)
			os.Exit(7)
		}
		defer fin.Close()
	}
```

设置输出流，若缺省，则通过标准输出（屏幕或重定向）读取输入流。并通过 StdinPipe 建立连接到 cmd 标准输入的管道，将当前的输出作为 cmd 的输入。

```go
	var fout io.WriteCloser
	cmd := &exec.Cmd{}

	if len(sa.printDest) == 0 {
		fout = os.Stdout
	} else {
		cmd = exec.Command("cat")
		// 用只写的方式打开 print_dest 文件，如果文件不存在，就创建该文件。
		var outputErr error
		cmd.Stdout, outputErr = os.OpenFile(sa.printDest, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open file %s\n",
				progname, sa.printDest)
			os.Exit(8)
		}

		// StdinPipe返回一个连接到command标准输入的管道pipe
		fout, outputErr = cmd.StdinPipe()
		if outputErr != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open pipe to \"lp -d%s\"\n",
				progname, sa.printDest)
			os.Exit(8)
		}

		cmd.Start()
		defer fout.Close()
	}
```
打印，根据 pageType 选择按固定行数打印或分页符（在这里用换行符替代）打印。

按固定行数打印：
```go
		line := 0
		page := 1
		for {
			line, crc := bufFin.ReadString('\n')
			if crc != nil {
				break 	// 读完一行
			}
			line++		// 行数加一
			if line > sa.pageLen { 	//读完一页
				page++	// 页数加一
				line = 1
			}
			// 到达指定页码，开始打印
			if (page >= sa.startPage) && (page <= sa.endPage) {
				_, err := fout.Write([]byte(line))
				if err != nil {
					fmt.Println(err)
					os.Exit(9)
				}
			}
		}
```

按分页符（换行符）打印：
```go
		page = 1
		for {
			page, err := bufFin.ReadString('\n')
			if err != nil {
				break // 读完一行
			}
			// 到达指定页码，开始打印
			if (page >= sa.startPage) && (page <= sa.endPage) {
				_, err := fout.Write([]byte(page))
				if err != nil {
					os.Exit(5)
				}
			}
			// 每碰到一个换页符都增加一页
			page++
		}
```
# 4、使用 selpg
① 该命令将把“input_file”的第 1 页写至标准输出（也就是屏幕），因为这里没有重定向或管道。
```
$ selpg -s1 -e1 input_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162013346.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

② 该命令与示例 1 所做的工作相同，但在本例中，selpg 读取标准输入，而标准输入已被 shell／内核重定向为来自“input_file”而不是显式命名的文件名参数。输入的第 1 页被写至屏幕。
```
$ selpg -s1 -e1 < input_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162740534.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

③ “other_command”的标准输出被 shell／内核重定向至 selpg 的标准输入。将第 10 页到第 20 页写至 selpg 的标准输出（屏幕）。
```
$ other_command | selpg -s10 -e20
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162752818.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

④ “other_command”的标准输出被 shell／内核重定向至 selpg 的标准输入。将第 10 页到第 20 页写至 selpg 的标准输出（屏幕）。
```
$ selpg -s10 -e20 input_file >output_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162816649.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

⑤ selpg 将第 10 页到第 20 页写至标准输出（屏幕）；所有的错误消息被 shell／内核重定向至“error_file”。请注意：在“2”和“>”之间不能有空格；这是 shell 语法的一部分（请参阅“man bash”或“man sh”）。
```
$ selpg -s10 -e20 input_file 2>error_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162830144.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

⑥ selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至“error_file”。当“input_file”很大时可使用这种调用；您不会想坐在那里等着 selpg 完成工作，并且您希望对输出和错误都进行保存。
```
$ selpg -s10 -e20 input_file >output_file 2>error_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/2019100416284234.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

⑦ selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至 /dev/null（空设备），这意味着错误消息被丢弃了。设备文件 /dev/null 废弃所有写至它的输出，当从该设备文件读取时，会立即返回 EOF。
```
$ selpg -s10 -e20 input_file >output_file 2>/dev/null
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162853598.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

⑧ selpg 将第 10 页到第 20 页写至标准输出，标准输出被丢弃；错误消息在屏幕出现。这可作为测试 selpg 的用途，此时您也许只想（对一些测试情况）检查错误消息，而不想看到正常输出。
```
$ selpg -s10 -e20 input_file >/dev/null
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/2019100416290962.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FpYW9femhhbmc=,size_16,color_FFFFFF,t_70)

⑨ selpg 的标准输出透明地被 shell／内核重定向，成为“other_command”的标准输入，第 10 页到第 20 页被写至该标准输入。“other_command”的示例可以是 lp，它使输出在系统缺省打印机上打印。“other_command”的示例也可以 wc，它会显示选定范围的页中包含的行数、字数和字符数。“other_command”可以是任何其它能从其标准输入读取的命令。错误消息仍在屏幕显示。
```
$ selpg -s10 -e20 input_file | other_command
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162932382.png)

⑩ 与上面的示例 9 相似，只有一点不同：错误消息被写至“error_file”。
```
$ selpg -s10 -e20 input_file 2>error_file | other_command
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162941786.png)

⑾ 该命令将页长设置为 66 行，这样 selpg 就可以把输入当作被定界为该长度的页那样处理。第 10 页到第 20 页被写至 selpg 的标准输出（屏幕）。
```
$ selpg -s10 -e20 -l66 input_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004162952299.png)

⑿ 该命令将页长设置为 66 行，这样 selpg 就可以把输入当作被定界为该长度的页那样处理。第 10 页到第 20 页被写至 selpg 的标准输出（屏幕）。
```
$ selpg -s10 -e20 -f input_file
```
测试结果如图：

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191004163001614.png)
