package gounit

//Request is a JSON object that is being read from Stdin
//in JSON mode (see -json command line flag)
type Request struct {
	InputFilePath  string `json:"inputFilePath"`
	OutputFilePath string `json:"outputFilePath"`
	InputFile      string `json:"inputFile"`
	OutputFile     string `json:"outputFile"`
	Comment        string `json:"comment"`
	Lines          []int  `json:"lines"`
}

//Response is an JSON object that is written to stdout
//in JSON mode
type Response struct {
	GeneratedCode string `json:"generatedCode"`
}
