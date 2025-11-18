### 手搓&收集整理的一些常用函数

#### -----赋值类(assign.go)-----

//给指针变量赋值, 可自动类型转换赋值（注：第1个为目标参数()，须传入指针方能赋值成功）
func SetValue(dPtr interface{}, s interface{}) 
<details>
<summary><b>示例</b></summary>
{示例代码}
</details>

- 结构体赋值 

//获取结构体指定名称Field字段的值（注：o结构体可以是指针也可是实参）
func GetField(o interface{}, name string) *Metadata 

//将结构体指定Field名称字段的值赋给指定变量（注：o结构体可以是指针也可是实参, d目标变量必须是指针）
func AssignFieldValue(o interface{}, name string, dPtr interface{}) error 

//设置结构体指定Field字段的值，可自动类型转换(注：结构体只能以指针传入，否则赋值失败)
func SetFieldValue(oPtr interface{}, name string, val interface{}) error 

//同上SetFieldValue，字段名字忽略大小写
func SetFieldVariant(oPtr interface{}, name string, val interface{}) error 


//给目标用任意类型进行赋值
func AssignVal(dt, st reflect.Type, dv, sv reflect.Value) 


#### -----转换类(convert.go)-----

//字符串外加单引号(原串中的单引号将被替换为2个单引号)
func QuotedStr(s string) string 

//前置补0
func ZeroPrefix(val, i int) string 
func ZeroRepeat(s string, i int) string 

//整数转16进制串
func IntToHex(val, size int) string

//16进制串转为整数
func HexToInt(str string, size int) int 

//字符串转换为int整数,出错返回默认值，无默认值则返回0
func StrToInt(s string, d ...int) int 

//字符串转换为float64小数,出错返回默认值，无默认值则返回0.0
func StrToFloat(s string, d ...float64) float64 

// Round 四舍五入，precision（十进制小数点后数字的数目）进行四舍五入的结果。precision 也可以是负数或零。
func Round(val float64, precision int) float64 

//取整数部份
func Trunc(f float64) int 

//取小数部份
func Frac(f float64) float64 

//去除字符串两端无用字符集chars_
func Strip(s_ string, chars_ string) string {

//利用正则表达式压缩字符串，去除空格或制表符
func CompressStr(str string) string {

// 任意类型数据转换为下列指定类型：
func Byte(i interface{}) byte 
func Bytes(i interface{}) []byte 
func Rune(i interface{}) rune 
func Runes(i interface{}) []rune 
func String(i interface{}) string 
func Bool(i interface{}) bool 	//if <i> is: false, "", 0, "false", "off", "no", empty slice/map.
func Int(i interface{}) int		//... int8,int16,int32,int64
func Uint(i interface{}) uint 	//... uint8,uint16,uint32,uint64
func Float32(i interface{}) float32 
func Float64(i interface{}) float64 
// 以上功能也可以用Metadata对变量进行包裹实现：Metadata{any}.Byte()/Rune()/String()/Bool()...。

//字符串转为[]byte
func Str2bytes(s string) []byte {

//[]byte转换为字符串
func Bytes2str(b []byte) string {

//结构转换为Map
func Struct2Map(that interface{}) map[string]interface{} {

//将数组格式化为字符串
func ArrayToString(array []interface{}) string {

//数字转换为中文金额串,保留2位小数
func NumToCNMoney(p_money float64) string {

- json转换(json.go)

// json串转换为map
func JsonToMap(jsonStr string) (map[string]interface{}, error) 

// map转换为json串
func MapToJson(m map[string]interface{}) (string, error) 

#### -----日期时间类(dateutils.go)-----

//当前日期
func Today() time.Time 

//格式化日期时间(改123456为ymdhns)
func FormatDateTime(format string, ts ...time.Time) string 

//字符串转换为时间(自动尝试各种格式)
func StrToDateTime(value string) time.Time 

func StrToLocalTime(value string) time.Time 

//日期时间转为字符串
func DatetimeToStr(d time.Time) string {

//日期转为字符串
func DateToStr(d time.Time) string {

//时间转为字符串
func TimeToStr(d time.Time) string {

// 获取当前时间戳（秒）
func NowTimestamp() int64 {

// 时间戳转化为当前中国时间 GMT，返回string
// 自己指定日期格式 例如 2006-01-02 15:04:05
func TimestampToChina(timestamp int64, format string) string {

// 中国时间字符串转化为时间戳
func ChinaToTimestamp(tmChina string, format string) int64 {

//2个时间(t1-t2)相差的天数
func TimeSubDay(t1, t2 time.Time) int {

//时间加减durationString（ns,us,ms,s,m,h） 加1天: 24h 减1天: -24h; 299ms, -1.5h, 2h40m
func TimeAdd(t time.Time, duration string, iMul ...time.Duration) time.Time {


#### -----工具类(utils.go)-----
	
//忽略大小写的字符串替换(替换所有)
func StringReplaceAll(s, old, new string) string {

//删除串中指定长度
func Delete(s *string, start, length int) 

//取子串
func SubStr(s string, start, length int) string 


//返回(UTF-8)子串sub在字符串str中的起始位置
func Pos(str, sub string) (index int) 

//取子串(UTF-8)从0开始，length大于串尾或=0取余下全部；start<0从尾部开始，length<0向前取
func Copy(s string, start, length int) string 


// 非空校验
func isBlank(value reflect.Value) bool 

//判断 interface 是否为 nil
func IsNil(x interface{}) bool 

//是否是整数字符串
func IsInt(s string) bool 

//是否是数字串(整数、小数)
func IsNum(s string) bool 

//字符串中是否有中文
func HasChineseChar(s string) bool


//In功能
func In(haystack []int, needle int) bool 

//支持像解释语言一样的通用 in 功能
func Contains(haystack interface{}, needle interface{}) (bool, error) 

//整除
func Div(a, b int) int 

//计算X的Y次方（X^Y)
func PowInt(x, y int) int {

//iif
func If(b bool, t, f interface{}) interface{} {

//查找数组、切片、Map中是指定数据索引号(无返-1；Map返0)s
func IndexOf(in interface{}, v interface{}) int {
	
//查找[]map[key]value列表内指定key及value在数组、切片列表中的index
func LocateMapList(in interface{}, key interface{}, v interface{}) int {


#### -----文件类(appfile.go)-----

//当前项目根目录(当前执行文件在根目录的子目录下时)
func RootDir() string 

//App文件名（无路径）
func AppFileName() string {

//App程序名（无文件后缀）
func AppName() string {

//App全路径文件名
func AppPath() string {

//App文件目录名
func AppDir() string {

//判断文件或文件夹是否存在
func IsExist(path string) bool

// return length in bytes for regular files
func FileSize(file string) int64

// joinFilePath joins path & file into a single string
func JoinFilePath(path, file string) string

function ChangeFileExt(const FileName, Extension: string): string;
function ChangeFilePath(const FileName, Path: string): string;
function ExtractFilePath(const FileName: string): string;
function ExtractFileDir(const FileName: string): string;
function ExtractFileName(const FileName: string): string;
function ExtractFileExt(const FileName: string): string;


#### -----网络类(net.go)-----

//返回本机IP地址串
func LocalIP() string 

// RemoteIp 返回远程客户端的 IP，如 192.168.1.1
func RemoteIp(req *http.Request) string 

// Ip2long 将 IPv4 字符串形式转为 uint32
func Ip2long(ipstr string) uint32 

//发送GET请求
//url:请求地址
//response:请求返回的内容
func Get(url string) (response string) 

//发送POST请求
//url:请求地址，data:POST请求提交的数据,contentType:请求体格式，如：application/json
//content:请求放回的内容
func Post(url string, data interface{}, contentType string) (content string) 
