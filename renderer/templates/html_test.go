package templates

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeUrisAbsolute(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><head></head><body>
<p><a href="relative.html">Relative</a></p>
<p><a href="/site-absolute.html">Site absolute</a></p>
<p><a href="https://absolute/path.html">Absolute</a></p>
<img src="relative.jpg"/>
<img src="/site-absolute.jpg"/>
<img src="https://absolute/image.jpg"/>
</body></html>`)

	w := strings.Builder{}

	err := MakeUrisAbsolute(r, &w, "http://site/base/", "path/index.html")
	if err != nil {
		panic(err)
	}

	expected := `<!DOCTYPE html><html><head></head><body>
<p><a href="http://site/base/path/relative.html">Relative</a></p>
<p><a href="http://site/base/site-absolute.html">Site absolute</a></p>
<p><a href="https://absolute/path.html">Absolute</a></p>
<img src="http://site/base/path/relative.jpg"/>
<img src="http://site/base/site-absolute.jpg"/>
<img src="https://absolute/image.jpg"/>
</body></html>`
	assert.Equal(t, expected, w.String())
}

func TestConvertToTextLineBreaks(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><body>
<p>En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín flaco y galgo corredor.</p>
<p>Una olla de algo más vaca que carnero, salpicón las más noches, duelos y quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los domingos, consumían las tres partes de su hacienda.</p><p>El resto della concluían sayo de velarte, calzas de velludo para las fiestas, con sus pantuflos de lo mesmo, y los días de entresemana se honraba con su vellorí de lo más fino.</p>

<p>Tenía en su casa una ama que pasaba de los cuarenta,</p>
<p>    y una sobrina que no llegaba a los veinte,
y un mozo de campo y plaza, que así    ensillaba el rocín como tomaba la podadera.   </p>
<p>Frisaba la edad de nuestro hidalgo conloscincuentaañoseradecomplexiónreciasecodecarnesenjutoderostrogranmadrugadoryamigodelacaza.
Quieren decir que tenía el sobrenombre de Quijada, o Quesada, que en esto hay alguna​diferencia​en​los​autores​que​deste​caso​escriben;​aunque​por​conjeturas​verosímiles​se​deja​entender​que​se​llamaba​Quijana.</p>
<p>Pero esto importa poco a nuestro cuento: basta que en la narración dél no se salga un punto de la verdad.</p>
</body></html>`)

	w := strings.Builder{}

	err := HtmlToText(r, &w, "Links", "Picture")
	if err != nil {
		panic(err)
	}

	expected := `En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.

Una olla de algo más vaca que carnero, salpicón las más noches, duelos y
quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los
domingos, consumían las tres partes de su hacienda.

El resto della concluían sayo de velarte, calzas de velludo para las fiestas,
con sus pantuflos de lo mesmo, y los días de entresemana se honraba con su
vellorí de lo más fino.

Tenía en su casa una ama que pasaba de los cuarenta,

y una sobrina que no llegaba a los veinte, y un mozo de campo y plaza, que así
ensillaba el rocín como tomaba la podadera.

Frisaba la edad de nuestro hidalgo
conloscincuentaañoseradecomplexiónreciasecodecarnesenjutoderostrogranmadrugadoryamigodelacaza.
Quieren decir que tenía el sobrenombre de Quijada, o Quesada, que en esto hay
algunadiferenciaenlosautoresquedestecasoescriben;aunqueporconjeturas
verosímilessedejaentenderquesellamabaQuijana.

Pero esto importa poco a nuestro cuento: basta que en la narración dél no se
salga un punto de la verdad.`
	assert.Equal(t, expected, w.String())
}

func TestConvertToTextFormatting(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><body>
<p>En un lugar de <b>la Mancha</b>, de cuyo nombre no quiero acordarme,
no ha mucho tiempo que vivía un hidalgo de los de lanza en astillero,
adarga antigua, rocín flaco y galgo corredor.</p>
<p>Una olla de algo más vaca que carnero, salpicón las más noches, <i>duelos
y quebrantos</i> los sábados, lantejas los viernes, <span>algún palomino de
añadidura los domingos</span>, consumían las tres partes de su hacienda.</p>
</body></html>`)

	w := strings.Builder{}

	err := HtmlToText(r, &w, "Links", "Picture")
	if err != nil {
		panic(err)
	}

	expected := `En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.

Una olla de algo más vaca que carnero, salpicón las más noches, duelos y
quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los
domingos, consumían las tres partes de su hacienda.`
	assert.Equal(t, expected, w.String())
}

func TestConvertToTextPreformatted(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><body>
<p>En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.</p>

<pre>Una olla de algo <b>más vaca que carnero</b>,
salpicón las más noches, duelos y quebrantos los sábados, lantejas los viernes, algún palomino de añadidura <em>los domingos,
consumían</em> las tres partes de su hacienda.

El resto della concluían sayo de velarte,    calzas de velludo para las fiestas,
con sus pantuflos de lo mesmo, y los días de entresemana se honraba con su vellorí de lo más fino.</pre>

<p>Tenía en su casa una ama que pasaba de los cuarenta,</p>
</body></html>`)

	w := strings.Builder{}

	err := HtmlToText(r, &w, "Links", "Picture")
	if err != nil {
		panic(err)
	}

	expected := `En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.

Una olla de algo más vaca que carnero,
salpicón las más noches, duelos y quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los domingos,
consumían las tres partes de su hacienda.

El resto della concluían sayo de velarte,    calzas de velludo para las fiestas,
con sus pantuflos de lo mesmo, y los días de entresemana se honraba con su vellorí de lo más fino.

Tenía en su casa una ama que pasaba de los cuarenta,`
	assert.Equal(t, expected, w.String())
}

func TestConvertToTextLinks(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><body>
<p>En un lugar de la Mancha, <a href="link1.html">de cuyo nombre</a> no quiero
acordarme, no ha mucho tiempo que vivía un hidalgo de los de lanza en astillero,
adarga antigua, rocín flaco y galgo corredor.</p>
<p>Una olla de algo <a href="link2.html">más vaca
que carnero</a>, salpicón las más noches, duelos y quebrantos los sábados,
lantejas los viernes, algún palomino de añadidura los domingos, consumían las
tres partes de su hacienda.</p>
</body></html>`)

	w := strings.Builder{}

	err := HtmlToText(r, &w, "Links", "Picture")
	if err != nil {
		panic(err)
	}

	expected := `En un lugar de la Mancha, de cuyo nombre[1] no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.

Una olla de algo más vaca que carnero[2], salpicón las más noches, duelos y
quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los
domingos, consumían las tres partes de su hacienda.

Links:
  [1] link1.html
  [2] link2.html`
	assert.Equal(t, expected, w.String())
}

func TestConvertToTextPictures(t *testing.T) {
	r := strings.NewReader(`<!DOCTYPE html><html><body>
<p>En un lugar de la Mancha, de cuyo nombre no quiero acordarme,
no ha mucho tiempo que vivía un hidalgo de los de lanza en astillero,
adarga antigua, rocín flaco y galgo corredor.</p>
<img src="donquijote.jpg" alt="Don Quijote">
<p>Una olla de algo más vaca que carnero, <img src="olla.png" alt="Olla">
salpicón las más noches, duelos y quebrantos los sábados, lantejas los viernes,
algún palomino de añadidura los domingos, consumían las tres partes de su
hacienda.</p>
</body></html>`)

	w := strings.Builder{}

	err := HtmlToText(r, &w, "Links", "Picture")
	if err != nil {
		panic(err)
	}

	expected := `En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho
tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín
flaco y galgo corredor.

(Picture: Don Quijote)

Una olla de algo más vaca que carnero, (Picture: Olla) salpicón las más noches,
duelos y quebrantos los sábados, lantejas los viernes, algún palomino de
añadidura los domingos, consumían las tres partes de su hacienda.`
	assert.Equal(t, expected, w.String())
}
