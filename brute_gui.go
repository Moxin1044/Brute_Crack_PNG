package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"hash/crc32"
	"io/ioutil"
)

type GUI struct {
	window       fyne.Window
	fileEntry    *widget.Entry
	widthEntry   *widget.Entry
	heightEntry  *widget.Entry
	heightCheck  *widget.Check
	progress     *widget.ProgressBar
	actionButton *widget.Button
	saveButton   *widget.Button
	modifiedData []byte
	cancelFunc   context.CancelFunc
	isRunning    bool
}

func BruteForcePNG(ctx context.Context, filePath string, checkHeight bool, progressCallback func(float64)) (bool, []byte, int, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, nil, 0, err
	}

	if hex.EncodeToString(data[:8]) != "89504e470d0a1a0a" {
		return false, nil, 0, fmt.Errorf("无效的PNG文件头")
	}

	for i := 0; i < 0xFFFF; i++ {
		select {
		case <-ctx.Done():
			return false, nil, 0, nil
		default:
			progressCallback(float64(i) / 0xFFFF)
			modified := make([]byte, len(data))
			copy(modified, data)

			offset := 16
			if checkHeight {
				offset = 20
			}
			binary.BigEndian.PutUint32(modified[offset:], uint32(i))

			crc := crc32.ChecksumIEEE(modified[12:29])
			if crc == binary.BigEndian.Uint32(data[29:33]) {
				return true, modified, i, nil
			}
		}
	}
	return false, nil, 0, nil
}

func NewGUI() *GUI {
	myApp := app.New()
	window := myApp.NewWindow("PNG爆破工具 v3.0")
	window.Resize(fyne.NewSize(600, 400))

	actionBtn := widget.NewButton("开始爆破", nil)
	return &GUI{
		window:       window,
		fileEntry:    widget.NewEntry(),
		widthEntry:   widget.NewEntry(),
		heightEntry:  widget.NewEntry(),
		heightCheck:  widget.NewCheck("爆破高度模式", nil),
		progress:     widget.NewProgressBar(),
		actionButton: actionBtn,
		saveButton:   widget.NewButton("保存文件", nil),
	}
}

func (g *GUI) LoadUI() {
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "目标文件", Widget: container.NewBorder(nil, nil, nil,
				widget.NewButton("选择文件", g.SelectFile), g.fileEntry)},
			{Text: "当前宽度", Widget: g.widthEntry},
			{Text: "当前高度", Widget: g.heightEntry},
		},
	}

	g.progress.Min = 0
	g.progress.Max = 1
	g.saveButton.Disable()

	actionBar := container.NewHBox(
		g.actionButton,
		g.saveButton,
	)

	content := container.NewVBox(
		form,
		g.heightCheck,
		g.progress,
		actionBar,
	)

	g.window.SetContent(content)
}

func (g *GUI) SelectFile() {
	dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}

		path := reader.URI().Path()
		g.fileEntry.SetText(path)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}

		if len(data) < 24 {
			dialog.ShowError(fmt.Errorf("文件太小"), g.window)
			return
		}

		g.widthEntry.SetText(fmt.Sprintf("%d", binary.BigEndian.Uint32(data[16:20])))
		g.heightEntry.SetText(fmt.Sprintf("%d", binary.BigEndian.Uint32(data[20:24])))
	}, g.window).Show()
}
func (g *GUI) ToggleBruteForce() {
	if !g.isRunning {
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelFunc = cancel
		g.isRunning = true
		g.actionButton.SetText("停止爆破")
		g.saveButton.Disable()
		g.saveButton.Refresh() // 强制立即刷新
		var success bool
		var size int

		go func() {
			defer func() {
				g.isRunning = false
				g.actionButton.SetText("开始爆破")
				g.progress.SetValue(0)
				g.actionButton.Refresh()
				g.progress.Refresh()
			}()

			var err error
			success, g.modifiedData, size, err = BruteForcePNG(
				ctx,
				g.fileEntry.Text,
				g.heightCheck.Checked,
				func(p float64) {
					g.progress.SetValue(p)
					g.progress.Refresh()
				})

			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}

			if success {
				g.saveButton.Enable()
				g.saveButton.Refresh()
				dimType := "宽度"
				if g.heightCheck.Checked {
					dimType = "高度"
				}
				dialog.ShowInformation("成功", fmt.Sprintf("找到有效%s: %dpx", dimType, size), g.window)
			} else {
				dialog.ShowInformation("结果", "未找到匹配尺寸", g.window)
			}
		}()
	} else {
		if g.cancelFunc != nil {
			g.cancelFunc()
		}
		g.actionButton.SetText("开始爆破")
		g.progress.SetValue(0)
		g.saveButton.Disable()
		g.saveButton.Refresh()
	}
}

func (g *GUI) SaveFile() {
	if g.modifiedData == nil {
		dialog.ShowError(fmt.Errorf("请先进行爆破"), g.window)
		return
	}

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		if _, err := writer.Write(g.modifiedData); err != nil {
			dialog.ShowError(fmt.Errorf("保存失败: %v", err), g.window)
			return
		}

		dialog.ShowInformation("保存成功",
			"文件已保存至："+writer.URI().Path(),
			g.window)
	}, g.window)

	saveDialog.SetFileName("modified.png")
	saveDialog.Show()
}

func main() {
	gui := NewGUI()
	gui.LoadUI()
	gui.actionButton.OnTapped = gui.ToggleBruteForce
	gui.saveButton.OnTapped = gui.SaveFile
	gui.window.ShowAndRun()
}
