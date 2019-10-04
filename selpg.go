package main

/////////////////////////import////////////////////////////////

import (
	"bufio" // bufio 用来帮助处理 I/O 缓存
	"fmt"
	"io"
	"os"
	"os/exec"

	flag "github.com/spf13/pflag"
)

/////////////////////////selpg_args struct////////////////////////////////

type selpgArgs struct {
	startPage int // 开始页
	endPage   int // 结束页

	inFilename string // 输入文件名
	printDest  string // 输出文件名

	pageLen  int    // 每页的行数，默认为72
	pageType string // 'l'按行打印，'f'按换页符打印，默认按行
}

/////////////////////////global variable////////////////////////////////

var progname string // 保存名称（命令就是通过该名称被调用）的全局变量，作为在错误消息中显示之用

/////////////////////////main////////////////////////////////

func main() {
	sa := selpgArgs{}
	progname = os.Args[0]

	processArgs(&sa) // 处理参数
	processInput(sa) // 处理输入输出
}

/////////////////////////func process_args////////////////////////////////

func processArgs(sa *selpgArgs) {
	// 将flag绑定到sa的各个变量上
	flag.IntVarP(&sa.startPage, "start", "s", -1, "start page(>1)")
	flag.IntVarP(&sa.endPage, "end", "e", -1, "end page(>=start_page)")
	flag.IntVarP(&sa.pageLen, "len", "l", 10, "page len")
	flag.StringVarP(&sa.printDest, "dest", "d", "", "print dest")
	flag.StringVarP(&sa.pageType, "type", "f", "l", "'l' for lines-delimited, 'f' for form-feed-delimited. default is 'l'")
	flag.Lookup("type").NoOptDefVal = "f"

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"USAGE: \n%s -s start_page -e end_page [ -f | -l lines_per_page ]"+
				" [ -d dest ] [ in_filename ]\n", progname)
		flag.PrintDefaults()
	}

	flag.Parse()

	// 检查命令行合法性
	// os.Args是一个储存了所有参数的string数组，可以使用下标来访问参数

	if len(os.Args) < 3 { // 参数个数不够（至少为 progname -s start_page -e end_page）
		fmt.Fprintf(os.Stderr, "\n%s: not enough arguments\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	// 处理第一个参数 - start page
	// 第一个参数必须为's'，start_page 必须大于1，且小于计算机能表示的最大整数值
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

	// 处理第二个参数 - end page
	// 第一个参数必须为'e'，end_page 必须大于1，小于计算机能表示的最大整数值，且小于等于start_page
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

	// 处理page_len
	if sa.pageLen < 1 || sa.pageLen > (intMax-1) {
		fmt.Fprintf(os.Stderr, "\n%s: invalid page length %s\n", progname, sa.pageLen)
		flag.Usage()
		os.Exit(5)
	}

	// 处理in_filename
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
}

/////////////////////////func process_input////////////////////////////////

func processInput(sa selpgArgs) {
	// 输入流,输入可以来自终端（用户键盘），文件或另一个程序的输出
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
	bufFin := bufio.NewReader(fin) // 获取一个读取器变量

	// 输出流，输出可以是屏幕，文件或另一个文件的输入
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

	// 打印，根据page_type（按固定行数或分页符进行打印）

	var page int // 当前页数
	var line int // 当前行数

	if sa.pageType == "l" { // 按固定行数打印
		line := 0
		page := 1
		for {
			line, crc := bufFin.ReadString('\n')
			if crc != nil {
				break // 读完一行
			}
			line++                 // 行数加一
			if line > sa.pageLen { //读完一页
				page++ // 页数加一
				line = 1
			}
			// 到达指定页码，开始打印
			if (pageCtr >= sa.startPage) && (pageCtr <= sa.endPage) {
				_, err := fout.Write([]byte(line))
				if err != nil {
					fmt.Println(err)
					os.Exit(9)
				}
			}
		}
	} else { // 按分页符打印
		page = 1
		for {
			page, err := bufFin.ReadString('\n')
			// 使用\n代替换页符，而且便于测试
			// line, crc := bufFin.ReadString('\f')
			if err != nil {
				break // 读完一行
			}
			// 到达指定页码，开始打印
			if (pageCtr >= sa.startPage) && (pageCtr <= sa.endPage) {
				_, err := fout.Write([]byte(page))
				if err != nil {
					os.Exit(5)
				}
			}
			// 每碰到一个换页符都增加一页
			page++
		}
	}

	//if err := cmd.Wait(); err != nil {
	//handle err
	if page_ctr < sa.start_page {
		fmt.Fprintf(os.Stderr,
			"\n%s: start_page (%d) greater than total pages (%d),"+
				" no output written\n", progname, sa.start_page, page_ctr)
	} else if page_ctr < sa.end_page {
		fmt.Fprintf(os.Stderr, "\n%s: end_page (%d) greater than total pages (%d),"+
			" less output than expected\n", progname, sa.end_page, page_ctr)
	}
}
