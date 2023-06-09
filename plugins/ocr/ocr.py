# 引入库
import easyocr
import sys
image_url = sys.argv[1]
#设置要识别中文和英文两种语言
reader = easyocr.Reader(['ch_sim', 'en'], gpu=False) #如果不用gpu可以设gpu=False
#设置要识别的图片
result = reader.readtext(image_url, detail = 0)
#打印识别结果
print("\nresult===>" + " ".join(result))