package cnpj

import "testing"

func TestMapApiPJ(t *testing.T) {
	body := []byte(`{
	  "razao_social": "LEADS2B S/A",
	  "nome_fantasia": "LEADS2B.COM",
	  "matriz_filial": "MATRIZ",
	  "data_inicio_atividade": "2009-11-24",
	  "situacao_cadastral": {"descricao": "ATIVA"},
	  "cnae_principal": {"codigo": "6311900", "descricao": "Tratamento de dados"},
	  "cnaes_secundarios": [{"codigo": "8599604", "descricao": "Treinamento"}],
	  "endereco": {
	    "tipo_logradouro": "RUA",
	    "logradouro": "PADRE ANCHIETA",
	    "numero": "2285",
	    "complemento": "SALA 1",
	    "bairro": "BIGORRILHO",
	    "cep": "80730001",
	    "municipio": {"codigo_ibge": "4106902"}
	  },
	  "telefones": [{"ddd": "41", "numero": "30287828"}],
	  "email": "contabilidade@leads2b.com",
	  "simples": {"optante": false}
	}`)
	out, err := MapApiPJ(body)
	if err != nil {
		t.Fatal(err)
	}
	if out.NmEmpresarial != "LEADS2B S/A" || out.CdTipoEstabelecimento != "MATRIZ" || out.DtAbertura != "2009-11-24" {
		t.Fatalf("%+v", out)
	}
	if out.FlOptanteSimplesNacional == nil || *out.FlOptanteSimplesNacional {
		t.Fatal("simples")
	}
	if len(out.Atividades) != 2 || out.Atividades[0].CdAtividade != "6311900" || !out.Atividades[0].FlPrincipal {
		t.Fatalf("%+v", out.Atividades)
	}
	if out.Endereco.CdTipoLogradouro == nil || *out.Endereco.CdTipoLogradouro != "RUA" {
		t.Fatalf("tipo %+v", out.Endereco.CdTipoLogradouro)
	}
	if out.Endereco.DsLogradouro == nil || *out.Endereco.DsLogradouro != "PADRE ANCHIETA" {
		t.Fatalf("logradouro %+v", out.Endereco.DsLogradouro)
	}
	if out.Endereco.IdCidadeIbge == nil || *out.Endereco.IdCidadeIbge != 4106902 {
		t.Fatalf("ibge %+v", out.Endereco.IdCidadeIbge)
	}
	if len(out.Telefones) != 1 || out.Telefones[0].NrTelefone != "4130287828" {
		t.Fatalf("%+v", out.Telefones)
	}
}
