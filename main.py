import zlib


def png_crc32_crack_wh(file_name):
    with open(file_name, 'rb') as f:
        hexdata = f.read().hex()
    PNG_data = hexdata[:16]
    if PNG_data == "89504e470d0a1a0a":
        IHDR = bytes.fromhex(hexdata[24:32])
        width = int(hexdata[36:40], 16)
        height = int(hexdata[44:48], 16)
        str2 = bytes.fromhex(hexdata[48:58])
        crc32 = int(hexdata[58:66], 16)
        add_num = 20000  # 最大宽高，合理修改快速出flag
        for w in range(width, width + add_num):
            for h in range(height, height + add_num):
                width_bytes = w.to_bytes(4, 'big')
                height_bytes = h.to_bytes(4, 'big')
                if zlib.crc32(IHDR + width_bytes + height_bytes + str2) == crc32:
                    return f"PNG图片宽度：{ hex(w)} | {w}\nPNG图片高度：{hex(h)} | {h}"
            if zlib.crc32(IHDR + width_bytes + height_bytes + str2) == crc32:
                break
    else:
        return "可能不是PNG文件，或文件头有修改。\nPNG文件头：89504e470d0a1a0a"


if __name__ == "__main__":
    filename = "test.png"
    print(png_crc32_crack_wh(filename))
