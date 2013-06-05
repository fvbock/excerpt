package excerpt

import (
	"fmt"
	"testing"
)

func TestFindExcerptsBM(t *testing.T) {
	text := "search678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456in search----------------------------search in search"

	var l int = 50
	search := map[string]float64{"search": 1, "in": 0.05}
	bestBM := FindExcerptsBM(search, text, l, true, true)
	fmt.Println("DONE", bestBM)

	text2 := `「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。「なめ猫」生みの親、１億１０００万円脱税容疑

	読売新聞 5月16日(木)14時53分配信

	葬儀で稼いだ所得を隠し、約１億１０００万円を脱税したとして、葬儀会社「紫音」（東京都渋谷区）と「ファミリー共済会」（同豊島区）、ファミリー社前代表で両社の実質的経営者、津田覚(さとる)氏（６２）が、法人税法違反容疑で東京国税局から東京地検に告発されていたことがわかった。

	代理人の弁護士は、両社とも既に修正申告したとしている。

	弁護士や関係者の話によると、紫音は警視庁管内の約４０の警察署から、ファミリー社は病院から、それぞれ遺体搬送の依頼を受けた上、遺族の依頼で葬儀を行っていた。ところが一部の葬儀について売り上げと経費の両方を除外して葬儀が行われなかったように装うなどして、万円２０１１年までの３年間に計約３億８０００万円の法人所得を隠した疑いが持たれている。

	隠した所得は、津田氏のマンションやブランド品の購入費、預金、飲食費などに充てられていたという。

	津田氏は、１９８０年代初頭にブームとなった、不良学生風の格好をさせた猫のキャラクター商品「なめ猫」のプロデューサーで、現在も「なめ猫」商品の企画・管理会社を経営している。取材に対し、弁護士を通じて「取引先や関係各位に多大な迷惑をかけ、反省している」とコメントした。`

	// text2 = "万円23456789012345678901234567890123456789012345678万円-----万円万円------"

	var l2 int = 200
	search2 := map[string]float64{"万円": 1}
	bestBM = FindExcerptsBM(search2, text2, l2, true, true)
	fmt.Println("BEST:\n", bestBM, "\n")
}