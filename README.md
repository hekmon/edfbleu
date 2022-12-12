# Simulateur EDF Tarifs Bleu

Ce programme vous permets de calculer localement le prix de votre consommation sur les 3 différents tarifs réglémentés "bleu" d'EDF (base, heures creuses et tempo) à partir d'un export de vos données depuis le site d'Enedis.

## Utilisation

### Obtention de vos données

#### Prérequis

* Avoir un compteur Linky (sans cela vos données de consommation ne peuvent être mesurées et exportées).
* Avoir activé l'enregistrement et la collecte de la consommation horaire [ici](https://mon-compte-particulier.enedis.fr/donnees/).
* Attendre suffisement longtemps pour avoir une collection importante de données (idéalement 1 an pour le calcul tempo).

#### Export de vos données

Une fois la collecte activé et suffisamment de données disponibles, vous pouvez demander un export de vos données de consommation [ici](https://mon-compte-particulier.enedis.fr/suivi-de-mesures/?ajouter_telechargement=true). Faîtes bien attention à sélectionner `Consommation horaire` dans l'option types de données et à sélectionner une durée d'un an idéalement.
* Attendre quelques minutes (cela peut être plus ou moins long).
* Récupérer votre export [ici](https://mon-compte-particulier.enedis.fr/mes-telechargements-mesures/).

### Utilisation du programme

* Récupérer une version du programme compilé dans les [releases](https://github.com/hekmon/edfbleu/releases) ou le compiler vous même avec `go build`.
* Executer le programme en lui indiquant le chemin votre votre export de données Enedis (voir exemple ci dessous).

```plain
doudou@DGaming:~/go/edfbleu$ ./edfbleu -csv 'Enedis_Conso_Heure_20210901-20220831_XXXXXXXXXXXXXX.csv' -monthly
PRM:            XXXXXXXXXXXXXX
Start:          2021-08-31 00:00:00 +0200 CEST
End:            2022-09-01 00:00:00 +0200 CEST

* September 2021
Option base:    101.75€
Option HC:      102.72€
Option Tempo:   85.00€

* October 2021
Option base:    86.66€
Option HC:      87.50€
Option Tempo:   72.67€

* November 2021
Option base:    88.71€
Option HC:      89.36€
Option Tempo:   81.75€

* December 2021
Option base:    80.71€
Option HC:      81.02€
Option Tempo:   95.92€

* January 2022
Option base:    92.68€
Option HC:      93.38€
Option Tempo:   160.03€

* February 2022
Option base:    72.73€
Option HC:      73.00€
Option Tempo:   63.54€

* March 2022
Option base:    87.27€
Option HC:      87.81€
Option Tempo:   74.59€

* April 2022
Option base:    85.82€
Option HC:      86.52€
Option Tempo:   72.84€

* May 2022
Option base:    99.22€
Option HC:      99.85€
Option Tempo:   83.75€

* June 2022
Option base:    111.07€
Option HC:      111.99€
Option Tempo:   92.46€

* July 2022
Option base:    139.22€
Option HC:      141.24€
Option Tempo:   116.76€

* August 2022
Option base:    117.63€
Option HC:      118.61€
Option Tempo:   78.67€

* TOTAL
Option base:    1165.72€
Option HC:      1175.24€
Option Tempo:   1079.86€
doudou@DGaming:~/go/edfbleu$
```

## Notes

### Heures creuses

Différentes périodes d'heure creuses existent par commune (vous pouvez vérifiez [ici](https://www.enedis.fr/heures-creuses/standard)). Une commune peut d'ailleurs en avoir plusieurs différentes suivant les jours. Le calcul d'heures creuses de ce programme est donc une approximation, la période la plus répandue 23H30-7H30 a donc été sélectionnée pour les calculs. Gardez cependant en tête que votre commune a certainement des tranches horaires différentes et/ou dynamiques.

### Jours tempo

#### Jours rouges et blancs

Ce programme n'utilise pas (encore ?) d'API pour récupérer les couleurs des jours. Pour le moment chacun des jours est enregistré rétrospectivement dans le programme et ces dates sont compilées au démarrage. Cela veut dire que celui-ci doit être régulièrement mis à jour (surtout en hiver). Si votre export de données contient des dates qui ne sont pas connues par le programme vous aurez le warning suivant:

```plain
/!\ the data set contains values that are beyong the internal data this program has. Please update the code.
```

Si vous le rencontrez, n'hésitez pas à ouvrir une [issue](https://github.com/hekmon/edfbleu/issues) ou une [pull request](https://github.com/hekmon/edfbleu/pulls) afin de mettre à jour le programme.

#### Consommation sous tempo

Les valeurs que vous obtiendrez correspondent à votre consommation *passée* ou vous ne décaliez pas forcément votre consommation. Le but de tempo étant de vous encourager à décaller votre consommation en journée rouge (prix du kWh prohibitif), votre prix tempo sera certainement bien moins élevé si vous pouvez décaller vos consommations pendant ces plages horaires. Plus d'information sur tempo [ici](https://particulier.edf.fr/fr/accueil/gestion-contrat/options/tempo/details.html).
