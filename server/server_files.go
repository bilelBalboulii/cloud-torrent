package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"archive/zip"	
//	"github.com/jpillora/archive"
)

const fileNumberLimit = 1000

type fsNode struct {
	Name     string
	Size     int64
	Modified time.Time
	Children []*fsNode
}

func (s *Server) listFiles() *fsNode {
	rootDir := s.state.Config.DownloadDirectory
	root := &fsNode{}
	if info, err := os.Stat(rootDir); err == nil {
		if err := list(rootDir, info, root, new(int)); err != nil {
			log.Printf("File listing failed: %s", err)
		}
	}
	return root
}

func (s *Server) serveFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/download/") {
		url := strings.TrimPrefix(r.URL.Path, "/download/")
		//dldir is absolute
		dldir := s.state.Config.DownloadDirectory
		file := filepath.Join(dldir, url)
		//only allow fetches/deletes inside the dl dir
		if !strings.HasPrefix(file, dldir) || dldir == file {
			http.Error(w, "Nice try\n"+dldir+"\n"+file, http.StatusBadRequest)
			return
		}
		info, err := os.Stat(file)
		if err != nil {
			http.Error(w, "File stat error: "+err.Error(), http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			if info.IsDir() {
				 
				
				/*w.Header().Set("Content-Type", "application/zip")
				w.WriteHeader(200)
				//write .zip archive directly into response
				a := archive.NewZipWriter(w)
				a.AddDir(file)
				a.Close()
				*/
				 
 				    // Get a Buffer to Write To
				    outFile, err := os.Create(file+`.zip`)
				    if err != nil {
					fmt.Println(err)
				    }
				    defer outFile.Close()
 				    // Create a new zip archive.
				    w := zip.NewWriter(outFile)
 				    // Add some files to the archive.
				    addFiles(w, file, "")
 				    if err != nil {
					fmt.Println(err)
				    }
 				    // Make sure to check the error on Close.
				    err = w.Close()
				    if err != nil {
					fmt.Println(err)
				    }
				
				
				
				
				
				
				
				
				
				
				
			} else {
				f, err := os.Open(file)
				if err != nil {
					http.Error(w, "File open error: "+err.Error(), http.StatusBadRequest)
					return
				}
				http.ServeContent(w, r, info.Name(), info.ModTime(), f)
				f.Close()
			}
		case "DELETE":
			if err := os.RemoveAll(file); err != nil {
				http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	s.static.ServeHTTP(w, r)
}


func addFiles(w *zip.Writer, basePath, baseInZip string) {
    // Open the Directory
    files, err := ioutil.ReadDir(basePath)
    if err != nil {
        fmt.Println(err)
    }
     for _, file := range files {
        fmt.Println(basePath + file.Name())
        if !file.IsDir() {
            dat, err := ioutil.ReadFile(basePath + file.Name())
            if err != nil {
                fmt.Println(err)
            }
             // Add some files to the archive.
            f, err := w.Create(baseInZip + file.Name())
            if err != nil {
                fmt.Println(err)
            }
            _, err = f.Write(dat)
            if err != nil {
                fmt.Println(err)
            }
        } else if file.IsDir() {
             // Recurse
            newBase := basePath + file.Name() + "/"
            fmt.Println("Recursing and Adding SubDir: " + file.Name())
            fmt.Println("Recursing and Adding SubDir: " + newBase)
             addFiles(w, newBase, file.Name() + "/")
        }
    }
}



//custom directory walk

func list(path string, info os.FileInfo, node *fsNode, n *int) error {
	if (!info.IsDir() && !info.Mode().IsRegular()) || strings.HasPrefix(info.Name(), ".") {
		return errors.New("Non-regular file")
	}
	(*n)++
	if (*n) > fileNumberLimit {
		return errors.New("Over file limit") //limit number of files walked
	}
	node.Name = info.Name()
	node.Size = info.Size()
	node.Modified = info.ModTime()
	if !info.IsDir() {
		return nil
	}
	children, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("Failed to list files")
	}
	node.Size = 0
	for _, i := range children {
		c := &fsNode{}
		p := filepath.Join(path, i.Name())
		if err := list(p, i, c, n); err != nil {
			continue
		}
		node.Size += c.Size
		node.Children = append(node.Children, c)
	}
	return nil
}
