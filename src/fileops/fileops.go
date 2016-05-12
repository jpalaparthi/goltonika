// file operations
package fileops

import (
	"os"
	"fmt"
	"time"
	)

func WriteToFile(filename,sub,details string,logat time.Time)(error){
	f,err:=os.OpenFile(filename+".txt", os.O_APPEND|os.O_CREATE, 0666)
	defer f.Close();
	text:= "Logged at:"+logat.String()+"\tSubject:"+sub+"\tDetails:"+details+"\n"
	f.WriteString(text)
    return err;
}

func GetCombDate(dt time.Time)string{
	d:=fmt.Sprintf("%s-%d-%d",dt.Month().String(),dt.Day(),dt.Year())
	return d}

func CatchAndLog(fname, sub string,logat time.Time,err error)bool{
	if(err!=nil){
	fmt.Println(err.Error());
	WriteToFile(fname,sub,err.Error(),logat)
	return false
	}
	return true
}