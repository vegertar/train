package simnet

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strconv"

	"github.com/mmcloughlin/globe"
)

// City contains basic properties about a city.
type City struct {
	ID       int
	Name     string
	Province string
	District string
}

func (c City) String() string {
	return fmt.Sprintf("%s%s%s(%d)", c.District, c.Province, c.Name, c.ID)
}

// ISP represents an ISP.
type ISP float64

func (i ISP) String() string {
	return fmt.Sprintf("ISP(%f)", float64(i))
}

// Point is a pair of city and ISP.
type Point struct {
	City City
	ISP  ISP
}

// NewPoint creates a point from a port integer.
func NewPoint(port int) Point {
	city := names[port%len(names)]
	isp := isps[port%len(isps)]
	return Point{cities[city], isp}
}

// NewPointFromCity creates a point from a city where specified by name.
func NewPointFromCity(city string) Point {
	return Point{City: cities[city]}
}

// Isolate returns a value of isolation of two points.
func Isolate(a, b Point) float64 {
	aX := float64(a.City.ID-minCityID) / float64(maxCityID-minCityID)
	aY := float64(a.ISP-minISP) / float64(maxISP-minISP)

	bX := float64(b.City.ID-minCityID) / float64(maxCityID-minCityID)
	bY := float64(b.ISP-minISP) / float64(maxISP-minISP)

	diffX, diffY := aX-bX, aY-bY
	return math.Sqrt(diffX*diffX + diffY*diffY)
}

// Affinity contains affinity values between two points in each direction.
type Affinity []struct {
	// A is the start point.
	A Point
	// B is the end point.
	B Point
	// PacketLoss is the packet loss in percent.
	PacketLoss int
}

// Draw renders a graph at given png file.
func (graph Affinity) Draw(png string, side int) error {
	g := globe.New()
	g.DrawGraticule(10.0)
	g.DrawCountryBoundaries()

	drew := make(map[string]bool)
	for _, s := range graph {
		radius := 0.02
		name := s.A.City.Name
		a, _ := GetLocation(name)
		if !drew[name] {
			if firstClassCores[name] {
				radius = 0.08
			} else if secondClassInfluxes[name] {
				radius = 0.05
			}
			g.DrawDot(a.Latitude, a.Longitude, radius)
			drew[name] = true
		}

		name = s.B.City.Name
		b, _ := GetLocation(name)
		if !drew[name] {
			if firstClassCores[name] {
				radius = 0.08
			} else if secondClassInfluxes[name] {
				radius = 0.05
			}
			g.DrawDot(b.Latitude, b.Longitude, radius)
			drew[name] = true
		}

		if s.PacketLoss < 100 {
			c := color.NRGBA{0x00, 0x64, 0x3c, 192}
			if firstClassCores[name] {
				c = color.NRGBA{255, 0, 0, 255}
			}
			g.DrawLine(
				a.Latitude, a.Longitude,
				b.Latitude, b.Longitude,
				globe.Color(c),
			)
		}
	}

	center, _ := GetLocation("北京市")
	g.CenterOn(center.Latitude, center.Longitude)
	return g.SavePNG(png, side)
}

// NewAffinity produces arbitrary affinity.
func NewAffinity(points []Point) Affinity {
	r := make(Affinity, 0, len(points))
	for _, a := range points {
		for _, b := range points {
			if a == b {
				continue
			}

			reachable := false

			if a.City.Province == b.City.Province && a.ISP == b.ISP {
				// All IDCs within same province and same ISP construct a connected graph.
				reachable = true
			}

			if a.City.Province != b.City.Province && a.ISP == b.ISP {
				// Existing an IDC on capital of down region can reach to an IDC on capital of up region with same ISP.
				if a.City.District == b.City.District && secondClassInfluxes[b.City.Name] {
					reachable = true
				} else if firstClassCores[b.City.Name] {
					reachable = true
				}
			}

			if a.City.Province == b.City.Province && a.ISP != b.ISP {
				// Existing an IDC can reach to one of the base ISPs for any ISP within same province. Base is the minimal set of ISPs that our management machine built on.
				reachable = baseISPs[int(b.ISP)]
			}

			var z struct {
				A          Point
				B          Point
				PacketLoss int
			}

			z.A = a
			z.B = b
			z.PacketLoss = 100
			if reachable {
				z.PacketLoss = int(Isolate(a, b) * 100)
			}
			r = append(r, z)
		}
	}
	return r
}

var (
	cities               = make(map[string]City)
	minCityID, maxCityID int
	names                []string
	firstClassCores      = map[string]bool{
		"北京市": true,
		"上海市": true,
		"广州市": true,
	}
	secondClassInfluxes = map[string]bool{
		"成都市": true,
		"武汉市": true,
		"西安市": true,
		"沈阳市": true,
	}

	isps           []ISP
	minISP, maxISP ISP
	ispMean        = 0
	ispStddev      = 1
	baseISPs       = make(map[int]bool)
)

func init() {
	var table map[string]interface{}
	if err := json.Unmarshal([]byte(chinaCity), &table); err != nil {
		panic(err)
	}

	var (
		districtIndex, provinceIndex int
		minCityID                    = int(math.MaxInt32)
	)
	for provinceName, v1 := range table {
		provinceIndex++
		for i, v2 := range v1.([]interface{}) {
			var city City
			city.Name = v2.(string)
			city.Province = provinceName
			switch provinceName {
			case "北京市", "天津市", "河北省", "山西省", "内蒙古自治区":
				city.District = "华北"
				districtIndex = 1
			case "河南省", "湖北省", "湖南省":
				city.District = "华中"
				districtIndex = 2
			case "广西壮族自治区", "广东省", "海南省":
				city.District = "华南"
				districtIndex = 3
			case "上海市", "山东省", "江苏省", "安徽省", "浙江省", "福建省", "江西省":
				city.District = "华东"
				districtIndex = 4
			case "新疆维吾尔自治区", "青海省", "甘肃省", "宁夏回族自治区", "陕西省":
				city.District = "西北"
				districtIndex = 5
			case "西藏自治区", "四川省", "重庆市", "贵州省", "云南省":
				city.District = "西南"
				districtIndex = 6
			case "黑龙江省", "吉林省", "辽宁省":
				city.District = "东北"
				districtIndex = 7
			default:
				continue
			}

			city.ID, _ = strconv.Atoi(fmt.Sprintf("%v%03v%03v", districtIndex, provinceIndex, i))
			if city.ID < minCityID {
				minCityID = city.ID
			}
			if city.ID > maxCityID {
				maxCityID = city.ID
			}

			cities[city.Name] = city
			names = append(names, city.Name)
		}
		sort.Strings(names)
	}

	// lets ISP follow normal distribution, so makes base ISPs easily
	baseISPs[ispMean] = true
	baseISPs[ispMean-ispStddev] = true
	baseISPs[ispMean+ispStddev] = true
	minISP = ISP(math.MaxFloat64)
	const numberOfISP = 100
	for i := 0; i < numberOfISP; i++ {
		isp := ISP(rand.NormFloat64()*float64(ispStddev) + float64(ispMean))
		isps = append(isps, isp)
		if isp < minISP {
			minISP = isp
		}
		if isp > maxISP {
			maxISP = isp
		}
	}
}

// chinaCity is a JSON string downloaded from https://raw.githubusercontent.com/modood/Administrative-divisions-of-China/master/dist/pc.json,
// which contains all China provinces and prefectural cities
const chinaCity = `
{
  "北京市": [
    "北京市"
  ],
  "天津市": [
    "天津市"
  ],
  "河北省": [
    "石家庄市",
    "唐山市",
    "秦皇岛市",
    "邯郸市",
    "邢台市",
    "保定市",
    "张家口市",
    "承德市",
    "沧州市",
    "廊坊市",
    "衡水市"
  ],
  "山西省": [
    "太原市",
    "大同市",
    "阳泉市",
    "长治市",
    "晋城市",
    "朔州市",
    "晋中市",
    "运城市",
    "忻州市",
    "临汾市",
    "吕梁市"
  ],
  "内蒙古自治区": [
    "呼和浩特市",
    "包头市",
    "乌海市",
    "赤峰市",
    "通辽市",
    "鄂尔多斯市",
    "呼伦贝尔市",
    "巴彦淖尔市",
    "乌兰察布市",
    "兴安盟",
    "锡林郭勒盟",
    "阿拉善盟"
  ],
  "辽宁省": [
    "沈阳市",
    "大连市",
    "鞍山市",
    "抚顺市",
    "本溪市",
    "丹东市",
    "锦州市",
    "营口市",
    "阜新市",
    "辽阳市",
    "盘锦市",
    "铁岭市",
    "朝阳市",
    "葫芦岛市"
  ],
  "吉林省": [
    "长春市",
    "吉林市",
    "四平市",
    "辽源市",
    "通化市",
    "白山市",
    "松原市",
    "白城市",
    "延边朝鲜族自治州"
  ],
  "黑龙江省": [
    "哈尔滨市",
    "齐齐哈尔市",
    "鸡西市",
    "鹤岗市",
    "双鸭山市",
    "大庆市",
    "伊春市",
    "佳木斯市",
    "七台河市",
    "牡丹江市",
    "黑河市",
    "绥化市",
    "大兴安岭地区"
  ],
  "上海市": [
    "上海市"
  ],
  "江苏省": [
    "南京市",
    "无锡市",
    "徐州市",
    "常州市",
    "苏州市",
    "南通市",
    "连云港市",
    "淮安市",
    "盐城市",
    "扬州市",
    "镇江市",
    "泰州市",
    "宿迁市"
  ],
  "浙江省": [
    "杭州市",
    "宁波市",
    "温州市",
    "嘉兴市",
    "湖州市",
    "绍兴市",
    "金华市",
    "衢州市",
    "舟山市",
    "台州市",
    "丽水市"
  ],
  "安徽省": [
    "合肥市",
    "芜湖市",
    "蚌埠市",
    "淮南市",
    "马鞍山市",
    "淮北市",
    "铜陵市",
    "安庆市",
    "黄山市",
    "滁州市",
    "阜阳市",
    "宿州市",
    "六安市",
    "亳州市",
    "池州市",
    "宣城市"
  ],
  "福建省": [
    "福州市",
    "厦门市",
    "莆田市",
    "三明市",
    "泉州市",
    "漳州市",
    "南平市",
    "龙岩市",
    "宁德市"
  ],
  "江西省": [
    "南昌市",
    "景德镇市",
    "萍乡市",
    "九江市",
    "新余市",
    "鹰潭市",
    "赣州市",
    "吉安市",
    "宜春市",
    "抚州市",
    "上饶市"
  ],
  "山东省": [
    "济南市",
    "青岛市",
    "淄博市",
    "枣庄市",
    "东营市",
    "烟台市",
    "潍坊市",
    "济宁市",
    "泰安市",
    "威海市",
    "日照市",
    "莱芜市",
    "临沂市",
    "德州市",
    "聊城市",
    "滨州市",
    "菏泽市"
  ],
  "河南省": [
    "郑州市",
    "开封市",
    "洛阳市",
    "平顶山市",
    "安阳市",
    "鹤壁市",
    "新乡市",
    "焦作市",
    "濮阳市",
    "许昌市",
    "漯河市",
    "三门峡市",
    "南阳市",
    "商丘市",
    "信阳市",
    "周口市",
    "驻马店市"
  ],
  "湖北省": [
    "武汉市",
    "黄石市",
    "十堰市",
    "宜昌市",
    "襄阳市",
    "鄂州市",
    "荆门市",
    "孝感市",
    "荆州市",
    "黄冈市",
    "咸宁市",
    "随州市",
    "恩施土家族苗族自治州"
  ],
  "湖南省": [
    "长沙市",
    "株洲市",
    "湘潭市",
    "衡阳市",
    "邵阳市",
    "岳阳市",
    "常德市",
    "张家界市",
    "益阳市",
    "郴州市",
    "永州市",
    "怀化市",
    "娄底市",
    "湘西土家族苗族自治州"
  ],
  "广东省": [
    "广州市",
    "韶关市",
    "深圳市",
    "珠海市",
    "汕头市",
    "佛山市",
    "江门市",
    "湛江市",
    "茂名市",
    "肇庆市",
    "惠州市",
    "梅州市",
    "汕尾市",
    "河源市",
    "阳江市",
    "清远市",
    "东莞市",
    "中山市",
    "潮州市",
    "揭阳市",
    "云浮市"
  ],
  "广西壮族自治区": [
    "南宁市",
    "柳州市",
    "桂林市",
    "梧州市",
    "北海市",
    "防城港市",
    "钦州市",
    "贵港市",
    "玉林市",
    "百色市",
    "贺州市",
    "河池市",
    "来宾市",
    "崇左市"
  ],
  "海南省": [
    "海口市",
    "三亚市",
    "三沙市",
    "儋州市"
  ],
  "重庆市": [
    "重庆市"
  ],
  "四川省": [
    "成都市",
    "自贡市",
    "攀枝花市",
    "泸州市",
    "德阳市",
    "绵阳市",
    "广元市",
    "遂宁市",
    "内江市",
    "乐山市",
    "南充市",
    "眉山市",
    "宜宾市",
    "广安市",
    "达州市",
    "雅安市",
    "巴中市",
    "资阳市",
    "阿坝藏族羌族自治州",
    "甘孜藏族自治州",
    "凉山彝族自治州"
  ],
  "贵州省": [
    "贵阳市",
    "六盘水市",
    "遵义市",
    "安顺市",
    "毕节市",
    "铜仁市",
    "黔西南布依族苗族自治州",
    "黔东南苗族侗族自治州",
    "黔南布依族苗族自治州"
  ],
  "云南省": [
    "昆明市",
    "曲靖市",
    "玉溪市",
    "保山市",
    "昭通市",
    "丽江市",
    "普洱市",
    "临沧市",
    "楚雄彝族自治州",
    "红河哈尼族彝族自治州",
    "文山壮族苗族自治州",
    "西双版纳傣族自治州",
    "大理白族自治州",
    "德宏傣族景颇族自治州",
    "怒江傈僳族自治州",
    "迪庆藏族自治州"
  ],
  "西藏自治区": [
    "拉萨市",
    "日喀则市",
    "昌都市",
    "林芝市",
    "山南市",
    "那曲地区",
    "阿里地区"
  ],
  "陕西省": [
    "西安市",
    "铜川市",
    "宝鸡市",
    "咸阳市",
    "渭南市",
    "延安市",
    "汉中市",
    "榆林市",
    "安康市",
    "商洛市"
  ],
  "甘肃省": [
    "兰州市",
    "嘉峪关市",
    "金昌市",
    "白银市",
    "天水市",
    "武威市",
    "张掖市",
    "平凉市",
    "酒泉市",
    "庆阳市",
    "定西市",
    "陇南市",
    "临夏回族自治州",
    "甘南藏族自治州"
  ],
  "青海省": [
    "西宁市",
    "海东市",
    "海北藏族自治州",
    "黄南藏族自治州",
    "海南藏族自治州",
    "果洛藏族自治州",
    "玉树藏族自治州",
    "海西蒙古族藏族自治州"
  ],
  "宁夏回族自治区": [
    "银川市",
    "石嘴山市",
    "吴忠市",
    "固原市",
    "中卫市"
  ],
  "新疆维吾尔自治区": [
    "乌鲁木齐市",
    "克拉玛依市",
    "吐鲁番市",
    "哈密市",
    "昌吉回族自治州",
    "博尔塔拉蒙古自治州",
    "巴音郭楞蒙古自治州",
    "阿克苏地区",
    "克孜勒苏柯尔克孜自治州",
    "喀什地区",
    "和田地区",
    "伊犁哈萨克自治州",
    "塔城地区",
    "阿勒泰地区"
  ],
  "台湾省": [],
  "香港特别行政区": [],
  "澳门特别行政区": []
}
`
