// arvhive/tar 实现了对tar档案里面文件的访问 （压缩和解压）
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ErrPrintf(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}

func main() {
	// err := tarZip("test.txt", "test.tar")
	// ErrPrintf(err)
	// err := tarUnzip("test.tar")
	// ErrPrintf(err)
	src := "logs.tar.gz"
	dst := ""

	// if err := MultiTar(src, dst); err != nil {
	// 	log.Fatal(err)
	// }
	if err := MultiUnTar(dst, src); err != nil {
		log.Fatal(err)
	}
}

// tarZip 打包单个文件
func tarZip(srcFile, dstFile string) error {
	// 创建目标文件
	fw, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer fw.Close()

	// 创建writer
	tw := tar.NewWriter(fw)
	defer func() {
		if err := tw.Close(); err != nil {
			ErrPrintf(err)
		}
	}()

	// 获取文件信息
	fi, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	// 文件头部信息
	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	// 写入writer头部信息
	if err = tw.WriteHeader(hdr); err != nil {
		return err
	}

	// 获取文件内容
	fr, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer fr.Close()

	written, err := io.Copy(tw, fr)
	if err != nil {
		return err
	}

	log.Printf("一共写入了%d个字符", written)
	return nil
}

// tar 解压单个文件
func tarUnzip(srcFile string) error {
	fr, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer fr.Close()

	tr := tar.NewReader(fr)

	for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		ErrPrintf(err)
		fi := hdr.FileInfo()

		// 创建一个文件
		f, err := os.Create(fi.Name())
		if err != nil {
			return err
		}

		in, err := io.Copy(f, tr)
		if err != nil {
			return err
		}

		log.Printf("解压 %s 文件到 %s 文件，总共写入 %d 字符", srcFile, f.Name(), in)

		os.Chmod(fi.Name(), fi.Mode().Perm())

		f.Close()
	}

	return nil
}

// MultiTar 打包多个文件
func MultiTar(src, dst string) (err error) {
	f, err := os.Create(dst)
	ErrPrintf(err)
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		h, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		// !!! 将 / 去掉
		h.Name = strings.TrimPrefix(fileName, string(filepath.Separator))

		if err := tw.WriteHeader(h); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(fileName)
		defer func() {
			if err = fr.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		if err != nil {
			return err
		}
		n, err := io.Copy(tw, fr)
		if err != nil {
			return err
		}

		log.Printf("打包文件%s到%s, 写入%d个字符", fileName, dst, n)

		return nil
	})
}

// MultiUnTar 解压文件
func MultiUnTar(dst, src string) (err error) {
	fi, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fi.Close()

	gr, err := gzip.NewReader(fi)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		dstFileDir := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if isExist := ExistDir(dstFileDir); !isExist {
				if err = os.MkdirAll(dstFileDir, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			n, err := io.Copy(f, tr)
			if err != nil {
				return err
			}

			// 将解压结果输出显示
			fmt.Printf("成功解压： %s , 共处理了 %d 个字符\n", dstFileDir, n)

			f.Close()
		}
	}
}

func ExistDir(dst string) bool {
	fi, err := os.Stat(dst)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}
