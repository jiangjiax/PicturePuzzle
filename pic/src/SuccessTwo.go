package main

import (
	"flag"
	"strconv"
	"time"
	"io/ioutil"
	"github.com/astaxie/beego"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/draw"
	"os"
	"bufio"
	"image/png"
	"golang.org/x/image/font"
	"fmt"
	"github.com/nfnt/resize"
	"log"
			"image/jpeg"
		"strings"
	"io"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "static/fonts/msyh.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 20, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
)

//文字转图片
//text:文字,times:图片名字,width:宽,height:高,fontsize:字体大小,newcolor:字体颜色,rgbaColor:背景颜色(默认透明),btx:字x轴位置
func SetFontImg(text []string, times string, width int, height int, fontsize string, newcolor color.RGBA, rgbaColor color.RGBA, btx int) string {
	flag.Set("size", fontsize)
	if times == "" {
		times = strconv.FormatInt(time.Now().Unix(), 10)
	}
	flag.Parse()

	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		beego.Debug(err)
		return ""
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		beego.Debug(err)
		return ""
	}

	fg, bg := image.Black, image.Transparent
	if *wonb {
		fg, bg = image.White, image.Black
	}
	fg = image.NewUniform(newcolor)

	rgba := image.NewRGBA(image.Rect(0, 0, width, len(text)*height))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{rgbaColor}, image.ZP, draw.Over)
	draw.Draw(rgba, rgba.Bounds(), bg, image.Pt(0, 0), draw.Over)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	// Draw the text.
	pt := freetype.Pt(btx, int(c.PointToFixed(*size)>>6))
	for _, s := range text {
		_, err = c.DrawString(s, pt)
		if err != nil {
			beego.Debug(err)
			return ""
		}
		pt.Y += c.PointToFixed(*size * *spacing)
	}
	// Save that RGBA image to disk.
	outFile, err := os.Create("static/fontimg/" + times + ".png")
	if err != nil {
		beego.Debug(err)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		beego.Debug(err)
	}
	err = b.Flush()
	if err != nil {
		beego.Debug(err)
	}
	return "static/fontimg/" + times + ".png"
}

//按照number分隔字符串数组
func StrKnife(texts []string,number int) []string{
	textsLen := len([]rune(texts[0]))
	knife := 0
	for textsLen>number {
		newText := []rune(texts[knife])
		texts[knife] = string(newText[0:number])
		KnifeAfter := string(newText[number:textsLen])
		texts = append(texts,KnifeAfter)
		textsLen = len([]rune(texts[knife+1]))
		knife++
	}

	return texts
}

//拼图
func PicturePuzzle(BodyImgName string,img []string,x []int,y []int,bx int,by int,Colors color.RGBA){
	//背景图-空白
	body := `static/fontimg/` + BodyImgName + `.png`
	fileBody, err := os.Create(body)
	if err != nil {
		fmt.Println(err)
	}
	defer fileBody.Close()
	BodyPng := image.NewRGBA(image.Rect(0, 0, bx, by))
	rgba := Colors

	//draw
	draw.Draw(BodyPng, BodyPng.Bounds(), &image.Uniform{rgba}, image.ZP, draw.Over)
	for i:=0;i<len(img);i++ {
		// open
		file, err := os.Open(img[i])
		if err != nil {
			log.Fatal(err)
		}
		// decode
		imgAfter, err := png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		draw.Draw(BodyPng, BodyPng.Bounds(), imgAfter, imgAfter.Bounds().Min.Sub(image.Pt(x[i], y[i])), draw.Over)
	}
	png.Encode(fileBody, BodyPng)
}

//改变图片大小
func PictureDecode(BodyImgSrc string,width uint,height uint){
	// open
	file, err := os.Open(BodyImgSrc)
	if err != nil {
		log.Fatal(err)
	}
	// decode
	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	//change
	m := resize.Resize(height, width, img, resize.Lanczos3)
	out, err := os.Create("static/fontimg/a1s.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	png.Encode(out, m)
}

//获取图片宽高
func PictureXY(BodyImgSrc string) []int{
	// open
	file, err := os.Open(BodyImgSrc)
	if err != nil {
		log.Fatal(err)
	}
	// decode
	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y

	XY := []int{width,height}
	return XY
}

//jpg转png
func JpegToPng(src string){
	// open
	files, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	fm, fmerr := guessImageFormat(files)
	if fmerr != nil {
		log.Fatal(fmerr)
	}

	switch fm {
	case "png":
	case "jpeg":
		ProductSrc := src
		//open
		files, err := os.Open(ProductSrc)
		if err != nil {
			log.Fatal(err)
		}
		// decode
		imgs, err := jpeg.Decode(files)
		if err != nil {
			log.Fatal(err)
		}
		files.Close()

		out, _ := os.Create(src)
		defer out.Close()

		err = png.Encode(out, imgs)
		if err != nil {
			log.Println(err)
		}
	}
}

//猜格式
func guessImageFormat(r io.Reader) (format string, err error) {
	_, format, err = image.DecodeConfig(r)
	return
}

func Success()  {
	space := 2
	var spaceUnit uint = 2

	//1.获取该产品的产品图路径，并赋值给ProductSrc，这个图片必须是png格式的
	ProductSrc := "static/fontimg/a2.png"
	JpegToPng(ProductSrc)
	PictureDecode(ProductSrc,280*spaceUnit,280*spaceUnit)

	//2.获取该产品的标题，并赋值title
	title := "长期出售电解铜 好铜 量大价格从优 长期出售电解铜 好铜 量大价格从优"
	titleTexts := []string{title}
	titleNewTexts := StrKnife(titleTexts,21)
	titleColor1 := color.RGBA{0, 0, 0, 255}
	titleColor2 := color.RGBA{0, 0, 0, 0}
	titleSrc := SetFontImg(titleNewTexts, "title", 300*space, 22*space, strconv.Itoa(15*space), titleColor1,titleColor2, 10*space)
	titleXY := PictureXY(titleSrc)

	//3.获取该产品的类型是求购还是供应，并赋值cateId(求购0,供应1)
	cateId := "0"
	cate := "求购"
	cateColor2 := color.RGBA{255, 60, 60, 255}
	if cateId == "1" {
		cate = "供应"
		cateColor2 = color.RGBA{0, 255, 0, 255}
	}
	cateTexts := []string{cate}
	cateNewTexts := StrKnife(cateTexts,21)
	cateColor1 := color.RGBA{255, 255, 255, 255}
	cateSrc := SetFontImg(cateNewTexts, "cate", 40*space, 16*space, strconv.Itoa(12*space), cateColor1, cateColor2, 15)

	//4.获取该产品的价格，并赋值price;获取该产品的单位，并赋值unit
	price := "110.00"
	unit := "吨"
	prices := strings.Split(price, ".")
	price1 := prices[0]
	price1Texts := []string{price1}
	price1NewTexts := StrKnife(price1Texts,21)
	price1Color1 := color.RGBA{26, 42, 95, 255}
	price1Color2 := color.RGBA{0, 0, 0, 0}
	price1width := len([]rune(price1))
	price1Src := SetFontImg(price1NewTexts, "price", 22*space+15*space*(price1width-1), 25*space, strconv.Itoa(23*space), price1Color1, price1Color2, 10*space)
	price1XY := PictureXY(price1Src)
	price2 := "."+prices[1]+" / "+unit
	price2Texts := []string{price2}
	price2NewTexts := StrKnife(price2Texts,21)
	price2Color1 := color.RGBA{0, 0, 0, 255}
	price2Color2 := color.RGBA{0, 0, 0, 0}
	price2Src := SetFontImg(price2NewTexts, "price2", 300*space, 22*space, strconv.Itoa(15*space), price2Color1, price2Color2, 10*space)
	priceIconSrc := "static/fontimg/pay.png"
	PictureDecode(priceIconSrc,14*spaceUnit,14*spaceUnit)
	priceSrc := []string{priceIconSrc,price1Src,price2Src}
	pricex := []int{0*space,10*space,price1XY[0]}
	pricey := []int{9*space,0*space,8*space}
	PicturePuzzle("priceHigh",priceSrc,pricex,pricey,300*space,30*space,color.RGBA{255, 255, 255, 255})
	PSrc := `static/fontimg/priceHigh.png`

	//5.获取该产品的求购或供应数量，并赋值number2
	number1Texts := []string{cate}
	number1Color1 := color.RGBA{170, 170, 170, 255}
	number1Color2 := color.RGBA{0, 0, 0, 0}
	number1Src := SetFontImg(number1Texts, "number1", 33*space, 13*space, strconv.Itoa(12*space), number1Color1, number1Color2, 10*space)
	number2 := "535"
	number2Texts := []string{number2}
	number2Color1 := color.RGBA{255, 173, 64, 255}
	number2Color2 := color.RGBA{0, 0, 0, 0}
	number2width := len([]rune(number2))
	number2Src := SetFontImg(number2Texts, "number2", 17*space+7*space*number2width, 13*space, strconv.Itoa(12*space), number2Color1, number2Color2, 10*space)
	number2XY := PictureXY(number2Src)
	numberIconSrc := "static/fontimg/number.png"
	PictureDecode(numberIconSrc,14*spaceUnit,14*spaceUnit)
	number3Texts := []string{unit}
	number3Color1 := color.RGBA{170, 170, 170, 255}
	number3Color2 := color.RGBA{0, 0, 0, 0}
	number3Src := SetFontImg(number3Texts, "number3", 32*space, 13*space, strconv.Itoa(12*space), number3Color1, number3Color2, 10*space)
	numberSrc := []string{numberIconSrc,number1Src,number2Src,number3Src}
	numberx := []int{0*space,10*space,35*space,number2XY[0]+20*space}
	numbery := []int{0*space,0*space,0*space,0*space}
	PicturePuzzle("numberHigh",numberSrc,numberx,numbery,200*space,30*space,color.RGBA{255, 255, 255, 255})
	nSrc := `static/fontimg/numberHigh.png`

	//6.获取发布日期，赋值data;如果有截止日期，获取截止日期，赋值endTime截止日期, 如果没有，赋值endTime空字符串""
	data := "2019年10月10日发布"
	endTime := "，2019年10月10日截止"
	dataTexts := []string{data}
	dataColor1 := color.RGBA{170, 170, 170, 255}
	dataColor2 := color.RGBA{0, 0, 0, 0}
	datawidth := len([]rune(data))
	dataSrc := SetFontImg(dataTexts, "data", 28*space+8*space*datawidth, 13*space, strconv.Itoa(12*space), dataColor1, dataColor2, 10*space)
	timeIconSrc := "static/fontimg/time.png"
	PictureDecode(timeIconSrc,14*spaceUnit,14*spaceUnit)
	if endTime!="" {
		endTimeTexts := []string{endTime}
		endTimeColor1 := color.RGBA{170, 170, 170, 255}
		endTimeColor2 := color.RGBA{0, 0, 0, 0}
		endTimeSrc := SetFontImg(endTimeTexts, "endTime", 35*space+8*space*datawidth, 13*space, strconv.Itoa(12*space), endTimeColor1, endTimeColor2, 10*space)

		timeSrc := []string{timeIconSrc,dataSrc,endTimeSrc}
		timex := []int{0*space,10*space,125*space}
		timey := []int{0*space,0*space,0*space}
		PicturePuzzle("timeHigh",timeSrc,timex,timey,300*space,30*space,color.RGBA{255, 255, 255, 255})
	}else{
		timeSrc := []string{timeIconSrc,dataSrc}
		timex := []int{0*space,10*space}
		timey := []int{0*space,0*space}
		PicturePuzzle("timeHigh",timeSrc,timex,timey,300*space,30*space,color.RGBA{255, 255, 255, 255})
	}
	tSrc := `static/fontimg/timeHigh.png`

	//7.二维码路径，赋值twoSrc
	twoSrc := "static/fontimg/two.png"
	PictureDecode(twoSrc,100*spaceUnit,100*spaceUnit)

	BodyImgName := "success"
	Src := []string{ProductSrc, titleSrc, cateSrc, PSrc, nSrc, tSrc, twoSrc}
	x := []int{30*space, 19*space, 270*space, 30*space, 30*space, 30*space, 120*space}
	y := []int{30*space, 320*space, 30*space, titleXY[1]+325*space, titleXY[1]+360*space, titleXY[1]+385*space, titleXY[1]+430*space}
	PicturePuzzle(BodyImgName,Src,x,y,340*space,600*space,color.RGBA{255, 255, 255, 255})

	//success.png生成
}