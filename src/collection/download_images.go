package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	//"reflect"
)

const imgPath string = "./data/interim/images/"
const csvPath string = "./data/raw/inaturalist/observations-spain-full.csv"
const numberOfWorker int = 24
const urlIndex int = 12
const nameIndex int = 32             // Common name
const scientificNameIndex int = 31   // Scientific name
const reportRate = 100               // report progress every 100 download
var r, _ = regexp.Compile("/[0-9]+") // regex para strings extraer el ID de la foto. Modificado para el nuevo tipo de URLs

/* Lichen invalid families. They are erased as they introduce uncertainty on the model */
var notValidFamilies = []string{"Liquen", "liquen", "Cladonia", "Lobaria", "Xylodon", "Xanthoria", "Usnea", "Xanthoparmelia", "Xanthocarpia", "Xanthomendoza", "Xanthoriicola", "Candelariella", "Vulpicida", "Vuilleminia", "Variospora", "Urocystis", "Uromyces", "Ustilago", "Umbilicaria", "Trochila",
	"Triphragmium", "Tranzschelia", "Trachyspora", "Thelotrema", "Tilachlidium", "Titaeosporina", "Tephromela", "Teloschistaceae", "Teloschistes", "Taphrina", "Synchytrium", "Stereocaulon", "Stamnaria", "Squamarina", "Sphaerophorus", "Solorina", "Septoria", "Seifertia", "Sawadaea", "Rusavskia", "Ricasolia",
	"Rhytisma", "Rhopographus", "Rhizoplaca", "Rhizocarpon", "Ramularia", "Ramalina", "Pyrenodesmia", "Pycnothelia", "Punctelia", "Pucciniastrum", "Puccinia", "Psora", "Psilolechia", "Pseudomicrostroma", "Protoblastenia", "Protomyces", "Protoparmeliopsis", "Porpidia", "Polycauliona",
	"Podosphaera", "Pleurosticta", "Platismatia", "Placynthium", "Physconia", "Physcia", "Phyllosticta", "Phyllactinia", "Phragmotrichum", "Phragmidium", "Phlyctis", "Phleogena", "Phlebia", "Phellinidium", "Phaeophyscia", "Pertusaria", "Peridermium", "Peniophora", "Peltigera", "Pectenia",
	"Parmotrema", "Parmeliopsis", "Parmelina", "Ophioparma", "Ochropsora", "Ochrolechia", "Normandina", "Nephroma", "Neoerysiphe", "Neodasyscypha", "Myriolecis", "Myriolecis", "Mycoblastus", "Multiclavula", "Monilinia", "Microbotryum", "Meruliopsis", "Melanohalea", "Melanelixia", "Melampsoridium",
	"Melampsorella", "Melampsora", "Marchandiomyces", "Lyophyllum", "Lyomyces", "Lophodermium", "Lobothallia", "Lobarina", "Lichina", "Letharia", "Leptogium", "Lepraria", "Lepra", "Lecidella", "Lecidea", "Lecanora", "Lathagrium", "Lasiosphaeria", "Lasiobelonium", "Lasallia", "Kuehneola", "Kretzschmaria",
	"Jackrogersella", "Irpex", "Inonotusobliquus obliquus", "Imshaugia", "Illosporiopsis", "Icmadophila", "Hypoxylon", "Hypotrachyna", "Hyperphyscia", "Hydropisphaera", "Heterocephalacria", "Gymnosporangium", "Gyalolechia", "Graphis", "Golovinomyces", "Fomitiporia", "Flavoparmelia", "Flavocetraria",
	"Exobasidium", "Evernia", "Etheirodon", "Erythricium", "Erysiphe", "Epichloe", "Eocronartium", "Entyloma", "Endophyllum", "Enchylium", "Diploschistes", "Diploicia", "Diplocarpon", "Diatrype", "Datronia", "Cumminsiella", "Cudoniella", "Corticium", "Coniophora", "Colpoma", "Collema", "Coleosporium",
	"Coenogonium", "Clypeococcum", "Claviceps", "Circinaria", "Chrysothrix", "Chaenotheca", "Cetraria", "Caloplaca", "Calogaya", "Calicium", "Byssomerulius", "Buellia", "Bryoria", "Boeremia", "Blumeriella", "Biscogniauxia", "Biscogniauxia", "Basidioradulum", "Baeomyces", "Bacidia", "Athelia", "Athallia",
	"Aspicilia", "Arthonia", "Arrhenia", "Arctoparmelia", "Anaptychia", "Amyloporia", "Amandinea", "Alyxoria", "Alectoria", "Aecidium", "Acarospora", "fulgens", "johnsonii", "spathulata"}

type data struct {
	url  string
	name string
}

func main() {
	input := make(chan data, 10)
	output := make(chan bool, 10)
	var counter int = 0 // use to name downloaded image files
	var queue []data

	csvFile, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(csvFile)

	/* read csv file */
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		/* get rid of lichen from fungi dataset.  */
		if !stringContainsValueInList(line[nameIndex-1], notValidFamilies) {
			queue = append(queue, data{
				url:  line[urlIndex],
				name: line[nameIndex-1],
			})
		} else {
			fmt.Println("Deleted:", line[nameIndex-1])
		}
	}

	queue = queue[1:] // remove title row

	/* init worker */
	for i := 0; i < numberOfWorker; i++ {
		url := queue[0]
		queue = queue[1:]

		counter++
		input <- url
		go worker(input, output)
	}

	for i := 0; i < len(queue); i++ {
		data := queue[i]

		<-output
		input <- data
		go worker(input, output)

		if i%reportRate == 0 {
			fmt.Println(i)
		}
	}
}

func worker(input chan data, done chan bool) {
	data := <-input
	url := data.url
	name := data.name
	//fmt.Println(name)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		//fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	path := imgPath + name + "/"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	index := r.FindStringSubmatch(url) // Devuelve el match más a la izquierda y todos los submatches
	//fmt.Println(url)
	//fmt.Println(reflect.TypeOf(index))
	//fmt.Println(index[0][1:])

	file, err := os.Create(path + strings.Join(index, "") + ".jpg")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	done <- true
}

/* This function checks if some value of the string list is contained by the string */
func stringContainsValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.Contains(value, v) {
			return true
		}
	}
	return false
}
