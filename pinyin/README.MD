### 一组汉字首拼处理函数(包括五笔首码)

//{取字串拼音首码}
func GetHzPYM(Str string) string 

//{取字串五笔首码}
func GetHzWBM(Str string) string 

//{快速拼音比较} SearchMode: 0=全匹配 1=左匹配 2=like
func ComparePYM(Str, PyStr string, SearchMode int8) bool 

//{快速五笔比较} SearchMode: 0=全匹配 1=左匹配 2=like
func CompareWBM(Str, WbStr string, SearchMode int8) bool 


数字一、二...九 --> 可以用1,2...9进行匹配/过滤;

也可以进行方言模糊匹配：
var (
	DontDiff_H_F bool = false //不区分H/F
	DontDiff_L_N bool = false //不区分L/N
	DontDiff_R_L bool = false //不区分R/L(同ang用时)
	DontDiff_K_G bool = false //不区分K/G
)
