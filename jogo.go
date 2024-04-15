package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// Define os elementos do jogo
type Elemento struct {
	simbolo     rune
	cor         termbox.Attribute
	corFundo    termbox.Attribute
	tangivel    bool
	interagivel bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
	simbolo:     '☺',
	cor:         termbox.ColorBlack,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: false,
}

var inimigo = Elemento{
	simbolo:     '☠',
	cor:         termbox.ColorLightYellow,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: true,
}

var portal = Elemento{
	simbolo:     '⛩',
	cor:         termbox.ColorRed,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: true,
}

var cavalo = Elemento{
	simbolo:     '♞',
	cor:         termbox.ColorRed,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: true,
}

// Parede
var parede = Elemento{
	simbolo:     '▤',
	cor:         termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	corFundo:    termbox.ColorDarkGray,
	tangivel:    true,
	interagivel: false,
}

var objetivo = Elemento{
	simbolo:     '⛝',
	cor:         termbox.ColorDefault,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: true,
}

// Barrreira
var barreira = Elemento{
	simbolo:     '#',
	cor:         termbox.ColorRed,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: false,
}

// Vegetação
var vegetacao = Elemento{
	simbolo:     '♣',
	cor:         termbox.ColorGreen,
	corFundo:    termbox.ColorDefault,
	tangivel:    false,
	interagivel: false,
}

var chave = Elemento{
	simbolo:     '⚿',
	cor:         termbox.ColorYellow,
	corFundo:    termbox.ColorDefault,
	tangivel:    true,
	interagivel: true,
}

// Elemento vazio
var vazio = Elemento{
	simbolo:     ' ',
	cor:         termbox.ColorDefault,
	corFundo:    termbox.ColorDefault,
	tangivel:    false,
	interagivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
	simbolo:     '.',
	cor:         termbox.ColorDefault,
	corFundo:    termbox.ColorYellow,
	tangivel:    false,
	interagivel: false,
}

var mapa [][]Elemento
var posX, posY, posIX, posIY, posCX, posCY, p1X, p1Y, p2X, p2Y int
var ultimoElementoSobPersonagem = vazio
var ultimoElementoSobInimigo = vazio
var ultimoElementoSobCavalo = vazio
var statusMsg string
var fim = false
var montando = false
var derrotado = false
var key = false
var portalAberto = false
var win = false

var efeitoNeblina = true
var revelado [][]bool
var raioVisao int = 3

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("mapa.txt")
	if efeitoNeblina {
		revelarArea()
	}
	desenhaTudo()
	go moveInimigo()
	go moveCavalo()
	go ativaPortal()
	for {

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return // Sair do programa
			}
			if fim {
				continue
			}
			if ev.Ch == 'e' {
				interagir()
			} else {
				mover(ev.Ch)
				if efeitoNeblina {
					revelarArea()
				}
			}
			desenhaTudo()
		}
	}
}

func carregarMapa(nomeArquivo string) {
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	y := 0
	portais := 0
	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		var linhaRevelada []bool
		for x, char := range linhaTexto {
			elementoAtual := vazio
			switch char {
			case parede.simbolo:
				elementoAtual = parede
			case barreira.simbolo:
				elementoAtual = barreira
			case vegetacao.simbolo:
				elementoAtual = vegetacao
			case chave.simbolo:
				elementoAtual = chave
			case objetivo.simbolo:
				elementoAtual = objetivo
			case portal.simbolo:
				elementoAtual = portal
				if portais == 1 {
					p2X, p2Y = x, y
					portais++
				}
				if portais == 0 {
					p1X, p1Y = x, y
					portais++
				}
			case personagem.simbolo:
				// Atualiza a posição inicial do personagem
				posX, posY = x, y
				elementoAtual = vazio
			case inimigo.simbolo:
				posIX, posIY = x, y
				elementoAtual = vazio
			case cavalo.simbolo:
				posCX, posCY = x, y
				elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			linhaRevelada = append(linhaRevelada, false)
		}
		mapa = append(mapa, linhaElementos)
		revelado = append(revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range mapa {
		for x, elem := range linha {
			if efeitoNeblina == false || revelado[y][x] {
				termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
			} else {
				termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
			}
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func gameOver() {
	fim = true
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	gameOverMsg := "GAME OVER, você foi morto pelo inimigo."
	if win {
		gameOverMsg = "Parabéns, você venceu!!! ☺☺☺"
	}
	for i, c := range gameOverMsg {
		width, height := termbox.Size()
		x := width * 0
		y := height / 2
		termbox.SetCell(x+i, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	for i, c := range statusMsg {
		termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}

func revelarArea() {
	minX := max(0, posX-raioVisao)
	maxX := min(len(mapa[0])-1, posX+raioVisao)
	minY := max(0, posY-raioVisao/2)
	maxY := min(len(mapa)-1, posY+raioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Revela as células dentro do quadrado de visão
			revelado[y][x] = true
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mover(comando rune) {
	if fim {
		return
	}
	if montando {
		moveMontado(comando)
		return
	}
	dx, dy := 0, 0
	switch comando {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	novaPosX, novaPosY := posX+dx, posY+dy
	if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
		mapa[novaPosY][novaPosX].tangivel == false {
		mapa[posY][posX] = ultimoElementoSobPersonagem         // Restaura o elemento anterior
		ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
		posX, posY = novaPosX, novaPosY                        // Move o personagem
		mapa[posY][posX] = personagem
	}
}

func moveMontado(comando rune) {
	dx, dy := 0, 0
	switch comando {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	novaPosX, novaPosY := posX+dx, posY+dy
	if novaPosY >= 0 && novaPosY+1 < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) {
		mapa[posY][posX] = ultimoElementoSobPersonagem         // Restaura o elemento anterior
		ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
		if ultimoElementoSobPersonagem == cavalo {
			ultimoElementoSobPersonagem = vazio
		}
		posX, posY = novaPosX, novaPosY // Move o personagem
		mapa[posY][posX] = personagem
		novaPosCY := posY + 1
		novaPosCX := posX
		mapa[posCY][posCX] = ultimoElementoSobCavalo
		if comando == 's' {
			mapa[posCY][posCX] = personagem
			ultimoElementoSobPersonagem = ultimoElementoSobCavalo
		}
		ultimoElementoSobCavalo = mapa[novaPosCY][novaPosCX]
		posCX, posCY = novaPosCX, novaPosCY
		mapa[posCY][posCX] = cavalo

	}
}

func interagir() {
	if montando {
		statusMsg = fmt.Sprintf("Você parou de montar o cavalo")
		montando = false
		go moveCavalo()
		return
	}
	menorDistancia := 100.0
	x := 0
	y := 0
	for dx := -4; dx < 5; dx++ {
		for dy := -4; dy < 5; dy++ {
			distancia := math.Sqrt(float64(dx*dx + dy*dy))
			novaPosX, novaPosY := posX+dx, posY+dy
			if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) {
				aux := mapa[novaPosY][novaPosX]
				if aux.interagivel {
					if distancia < menorDistancia {
						menorDistancia = distancia
						x = posX + dx
						y = posY + dy
					}
				}
			}
		}
	}
	if menorDistancia < 100 {
		aux := mapa[y][x]
		char := aux.simbolo
		switch char {
		case cavalo.simbolo:
			if mapa[posCY+1][posCX].tangivel {
				statusMsg = fmt.Sprintf("Não é possível montar o cavalo aqui")
				return
			}
			novaPosX := posCX
			novaPosY := posCY - 1
			mapa[posY][posX] = ultimoElementoSobPersonagem
			ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
			posX, posY = novaPosX, novaPosY                        // Move o personagem
			mapa[posY][posX] = personagem                          // Coloca o personagem na nova posição
			montando = true
			statusMsg = fmt.Sprintf("Você está montando no cavalo!")
		case inimigo.simbolo:
			statusMsg = fmt.Sprintf("Você matou o inimigo em %d, %d", x, y)
			derrotado = true
		case chave.simbolo:
			statusMsg = fmt.Sprintf("Você achou uma chave escondida! Para que será que serve?")
			key = true
			mapa[y][x] = vazio
		case objetivo.simbolo:
			if !key {
				statusMsg = fmt.Sprintf("A parede parece diferente...")
			}
			if key {
				win = true
				gameOver()
			}
		case portal.simbolo:
			if !portalAberto {
				statusMsg = fmt.Sprintf("O que é isso, um templo chinês? um portal?")
			}
			if portalAberto {
				if int(y) == int(p1Y) {
					mapa[posY][posX] = ultimoElementoSobPersonagem
					ultimoElementoSobPersonagem = mapa[p2Y][p2X-2]
					posX, posY = p2X-3, p2Y
					mapa[posY][posX] = personagem
				}
				if int(y) == int(p2Y) {
					statusMsg = fmt.Sprintf("O portal te levou a um cômodo secreto...?")
					mapa[posY][posX] = ultimoElementoSobPersonagem
					ultimoElementoSobPersonagem = mapa[p1Y][p1X-2]
					posX, posY = p1X-3, p1Y
					mapa[posY][posX] = personagem
				}
			}
		}
	}
	if menorDistancia == 100 {
		statusMsg = fmt.Sprintf("não há ninguém para interagir")
	}
}

func moveInimigo() {
	for {
		n := rand.Intn(4)
		dx, dy := 0, 0
		switch n {
		case 0:
			dy = -1
		case 1:
			dx = -1
		case 2:
			dy = 1
		case 3:
			dx = 1
		}
		novaPosX, novaPosY := posIX+dx, posIY+dy
		if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
			mapa[novaPosY][novaPosX].tangivel == false {
			mapa[posIY][posIX] = ultimoElementoSobInimigo
			ultimoElementoSobInimigo = mapa[novaPosY][novaPosX]
			posIX, posIY = novaPosX, novaPosY
			mapa[posIY][posIX] = inimigo
		}
		time.Sleep(1 * time.Second)
		for dx := -1; dx < 2; dx++ {
			for dy := -1; dy < 2; dy++ {
				if mapa[posIY+dy][posIX+dx] == personagem {
					statusMsg = fmt.Sprintf("O inimigo te atacou em (%d, %d)", posX, posY)
					gameOver()
				}
			}
		}
		if fim {
			gameOver()
			return
		}
		if derrotado {
			mapa[posIY][posIX] = vazio
			return
		}
		desenhaTudo()
	}
}

func moveCavalo() {
	for {
		n := rand.Intn(8)
		dx, dy := 0, 0
		switch n {
		case 0:
			dy = +2
			dx = +1
		case 1:
			dx = +2
			dy = +1
		case 2:
			dy = -1
			dx = +2
		case 3:
			dx = +1
			dy = -2
		case 4:
			dy = +2
			dx = -1
		case 5:
			dx = -2
			dy = +1
		case 6:
			dy = -1
			dx = -2
		case 7:
			dx = -1
			dy = -2
		}
		novaPosX, novaPosY := posCX+dx, posCY+dy
		if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
			mapa[novaPosY][novaPosX].tangivel == false {
			mapa[posCY][posCX] = ultimoElementoSobCavalo
			ultimoElementoSobCavalo = mapa[novaPosY][novaPosX]
			posCX, posCY = novaPosX, novaPosY
			mapa[posCY][posCX] = cavalo
		}
		time.Sleep(1 * time.Second)
		if fim {
			gameOver()
			return
		}
		if montando {
			return
		}
		desenhaTudo()
	}
}

func ativaPortal() {
	for {
		if fim {
			gameOver()
			return
		}
		n := rand.Intn(10)
		portalAberto = true
		time.Sleep(time.Duration(n) * time.Second)
		portalAberto = false
		time.Sleep(4 * time.Second)
	}
}
