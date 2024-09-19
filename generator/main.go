package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/google/uuid"
)

var (
	names    []Name
	families []Family
)

type (
	Sex            int
	GenerationType int
)

const (
	GENERATIONS       = 10
	NAME_COUNT        = 200
	FAMILY_COUNT      = 100
	SEED              = 0.2
	COMPETITION_DELTA = 1.5
)

const (
	MALE = iota
	FEMALE
)

const (
	DECLINE = iota - 1
	STABLE
	GROWTH
)

type Family struct {
	id         string
	name       string
	generation int
}

type Person struct {
	mother     *Person
	father     *Person
	generation *Generation
	id         string
	name       string
	family     Family
	sex        Sex
}
type People []Person

func (s Sex) String() string {
	switch s {
	case 0:
		return "MALE"
	case 1:
		return "FEMALE"
	default:
		return "UNKNOWN"
	}
}

type Couple struct {
	mother *Person
	father *Person
	family Family
	lambda int
}
type Couples []Couple

type Generation struct {
	competition float64
	genType     GenerationType
	direction   GenerationType
	id          int
}

func (gt GenerationType) String() string {
	switch gt {
	case -1:
		return "DECLINE"
	case 0:
		return "STABLE"
	case 1:
		return "GROWTH"
	default:
		return "UNKNOWN"
	}
}

type History []Generation

type Name struct {
	value string
	sex   Sex
}

func readFirstNames() {
	file, err := os.Open("first_names.csv")
	if err != nil {
		fmt.Printf("failed to open first_names.csv: %v\n", err)
		panic(1)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read first_names.csv: %v\n", err)
		panic(1)
	}
	rows := strings.Split(string(content), "\n")

	for _, row := range rows {
		if len(row) == 0 {
			break
		}

		values := strings.Split(row, ",")

		var sex Sex
		if values[1] == "MALE" {
			sex = MALE
		} else {
			sex = FEMALE
		}

		names = append(names, Name{values[0], sex})
	}
}

func readLastNames() {
	file, err := os.Open("last_names.csv")
	if err != nil {
		fmt.Printf("failed to open last_names.csv: %v\n", err)
		panic(1)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read last_names.csv: %v\n", err)
		panic(1)
	}
	rows := strings.Split(string(content), "\n")

	for _, row := range rows {
		if len(row) == 0 {
			break
		}

		families = append(families, Family{uuid.NewString(), row, 0})
	}
}

func generateFirstName(s Sex) string {
	idx := rand.Intn(NAME_COUNT)
	if s == FEMALE {
		idx += NAME_COUNT
	}
	return names[idx].value
}

func generateFamily(g *Generation) Family {
	idx := rand.Intn(FAMILY_COUNT)
	return Family{uuid.NewString(), families[idx].name, g.id}
}

func generateLambda(g *Generation) int {
	l := 0
	switch g.genType {
	case DECLINE:
		l = 1
	case STABLE:
		l = 2
	case GROWTH:
		l = 4
	}

	return l
}

func createAdam(g Generation) Person {
	father := Person{}
	mother := Person{}
	family := generateFamily(&g)

	return Person{
		&mother,
		&father,
		&g,
		uuid.NewString(),
		generateFirstName(MALE),
		family,
		MALE,
	}
}

func createEve(g Generation) Person {
	father := Person{}
	mother := Person{}
	family := generateFamily(&g)

	return Person{
		&mother,
		&father,
		&g,
		uuid.NewString(),
		generateFirstName(FEMALE),
		family,
		FEMALE,
	}
}

func generatePoissonRV(lambda int) int {
	var n int

	for s := 0.0; s < 1; {
		u := rand.Float64()
		e := -math.Log(u) / float64(lambda)
		n += 1
		s += e
	}

	return n
}

func generateChildren(g *Generation, c Couple) People {
	var children People

	n := generatePoissonRV(c.lambda)
	for i := 0; i < n; i++ {
		sex := Sex(math.Round(rand.Float64()))
		child := Person{
			c.mother,
			c.father,
			g,
			uuid.NewString(),
			generateFirstName(sex),
			c.family,
			sex,
		}
		children = append(children, child)
	}

	return children
}

func simulate(gen Generation, males People, females People, everybody *People) (People, People) {
	var couples Couples
	var maxCouples int
	malesCount, femalesCount := len(males), len(females)
	if malesCount <= femalesCount {
		maxCouples = malesCount
	} else {
		maxCouples = femalesCount
	}

	rand.Shuffle(len(males), func(i int, j int) {
		males[i], males[j] = males[j], males[i]
	})
	rand.Shuffle(len(females), func(i int, j int) {
		females[i], females[j] = females[j], females[i]
	})
	for i := 0; i < maxCouples; i++ {
		couple := Couple{
			&females[i],
			&males[i],
			males[i].family,
			generateLambda(&gen),
		}
		couples = append(couples, couple)
	}

	var m, f People
	for _, c := range couples {
		children := generateChildren(&gen, c)
		for _, child := range children {
			*everybody = append(*everybody, child)
			switch child.sex {
			case MALE:
				m = append(m, child)
			case FEMALE:
				f = append(f, child)
			}
		}
	}

	return m, f
}

func incrementGeneration(g *Generation) Generation {
	var comp float64
	var dir GenerationType
	var genType GenerationType

	switch g.direction {
	case DECLINE:
		comp = math.Pow(g.competition, 1/COMPETITION_DELTA)
	case GROWTH:
		comp = math.Pow(g.competition, COMPETITION_DELTA)
	}

	if comp > 0.6 {
		genType = DECLINE
		dir = GROWTH
	} else if comp < 0.2 {
		genType = GROWTH
		dir = DECLINE
	} else {
		genType = STABLE
		dir = g.direction // previous direction
	}

	return Generation{comp, genType, dir, g.id + 1}
}

func export(history *History, everybody *People) {
	os.Remove("../imports")
	err := os.Mkdir("../imports", fs.FileMode(fs.ModeDir))

	// nodes
	fmt.Println("writing people.csv...")
	people, err := os.Create("../imports/people.csv")
	if err != nil {
		fmt.Printf("failed to create people.csv: %v\n", err)
		panic(1)
	}
	defer people.Close()

	buf := bufio.NewWriter(people)
	fmt.Fprintf(buf, "personId:ID,name:STRING,sex:STRING,motherId:STRING,fatherId:STRING,familyId:STRING,:LABEL\n")
	for _, p := range *everybody {
		fmt.Fprintf(buf, "%s,%s,%s,%s,%s,%s,%s\n", p.id, p.name, p.sex, p.mother.id, p.father.id, p.family.id, "person")
		buf.Flush()
	}

	fmt.Println("writing generations.csv...")
	generations, err := os.Create("../imports/generations.csv")
	if err != nil {
		fmt.Printf("failed to create generations.csv: %v\n", err)
		panic(1)
	}
	defer generations.Close()

	buf = bufio.NewWriter(generations)
	fmt.Fprintf(buf, "generationId:ID,Type:STRING,Direction:STRING,competitionFactor:FLOAT,:LABEL\n")
	for _, g := range *history {
		fmt.Fprintf(buf, "%d,%s,%s,%f,%s\n", g.id, g.genType, g.direction, g.competition, "Generation")
		buf.Flush()
	}

	// edges
	fmt.Println("writing edges.csv...")
	edges, err := os.Create("../imports/edges.csv")
	if err != nil {
		fmt.Printf("failed to create edges.csv: %v\n", err)
		panic(1)
	}
	defer edges.Close()

	buf = bufio.NewWriter(edges)
	fmt.Fprintf(buf, ":START_ID,:END_ID,:TYPE\n")
	for _, p := range *everybody {
		if p.father.id != "" {
			fmt.Fprintf(buf, "%s,%s,%s\n", p.id, p.father.id, "FATHER")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.id, p.father.id, "PARENT")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.father.id, p.id, "CHILD")
		}
		if p.mother.id != "" {
			fmt.Fprintf(buf, "%s,%s,%s\n", p.id, p.mother.id, "MOTHER")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.id, p.mother.id, "PARENT")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.mother.id, p.id, "CHILD")
		}
		fmt.Fprintf(buf, "%s,%d,%s\n", p.id, p.generation.id, "IN_GENERATION")
		buf.Flush()
	}

	fmt.Println("done!")
}

func main() {
	readFirstNames()
	readLastNames()

	generation := Generation{SEED, GROWTH, GROWTH, 0}
	adam := createAdam(generation)
	eve := createEve(generation)

	history := History{generation}
	males := People{adam}
	females := People{eve}
	everybody := People{adam, eve}

	for i := 0; i < GENERATIONS; i++ {
		generation = incrementGeneration(&generation)
		history = append(history, generation)

		fmt.Printf(
			"GENERATION: %d - TYPE: %s\nDir: %s - Competition Factor: %f\n",
			generation.id,
			generation.genType,
			generation.direction,
			generation.competition,
		)

		males, females = simulate(generation, males, females, &everybody)
		fmt.Printf(
			"Population: %d\nActive: %d\n\n",
			len(everybody),
			len(males)+len(females),
		)

		if len(males)+len(females) == 0 {
			fmt.Println("Family has died out :(")
			break
		}
	}
	fmt.Println()

	export(&history, &everybody)
}
