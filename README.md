# Simulateur EDF Tarifs Bleu

Ce programme vous permets de calculer localement le prix de votre consommation sur les 3 différents tarifs réglémentés "bleu" d'EDF (base, heures creuses et tempo) à partir d'un export de vos données depuis le site d'Enedis.

## Utilisation

### Obtention de vos données

#### Prérequis

* Avoir un compteur Linky (sans cela vos données de consommation ne peuvent être mesurées et exportées).
* Avoir activé l'enregistrement et la collecte de la consommation horaire [ici](https://mon-compte-particulier.enedis.fr/donnees/).
* Attendre suffisement longtemps pour avoir une collection importante de données (idéalement 1 an pour le calcul tempo).

#### Export de vos données

Une fois la collecte activé et suffisamment de données disponibles, vous pouvez demander un export de vos données de consommation [ici](https://mon-compte-particulier.enedis.fr/suivi-de-mesures/?ajouter_telechargement=true). Deux types d'exports sont disponibles, au format `.xlsx` :

* La **consommation quotidienne** (`Export_energie_...xlsx`) : la consommation totale de chaque jour, sur la période de votre choix en un seul fichier. Le détail intra-journée n'y figure cependant pas (voir la note correspondante plus bas).
* La **courbe de charge** (`Export_courbe_de_charge_...xlsx`) : la consommation par pas de 30 minutes, qui permet un calcul exact. Attention, Enedis limite ces exports à **7 jours par téléchargement** : pour couvrir une longue période, téléchargez donc plusieurs exports hebdomadaires, le programme sait les fusionner (voir plus bas).

Puis :
* Attendre quelques minutes (cela peut être plus ou moins long).
* Récupérer votre export [ici](https://mon-compte-particulier.enedis.fr/mes-telechargements-mesures/).

L'ancien export CSV horaire n'est plus proposé par Enedis mais reste supporté par le programme si vous en avez conservé.

### Utilisation du programme

* Récupérer une version du programme compilé dans les [releases](https://github.com/hekmon/edfbleu/releases) ou le compiler vous-même avec `go build`.
* Executer le programme en lui indiquant le chemin de votre export de données Enedis. Vous pouvez lui donner plusieurs fichiers à la suite pour les fusionner (pratique avec les exports hebdomadaires de courbe de charge), et l'ancien drapeau `-csv` fonctionne toujours pour indiquer un fichier.

Exemple avec deux exports hebdomadaires de courbe de charge (fichiers `.xlsx`) :

```plain
doudou@DGaming:~/go/edfbleu$ ./edfbleu -monthly courbe_de_charge_semaine1.xlsx courbe_de_charge_semaine2.xlsx
PRM:            12345678901234
Start:          2025-01-06 00:00:00 +0100 CET
End:            2025-01-20 00:00:00 +0100 CET
Stepping:       30m0s

* January 2025
Consommation:   ~224.15 kWh
Option base:    56.40€
Option HC:      57.02€
Option Tempo:   82.67€

* TOTAL
Consommation:   ~224.15 kWh
Option base:    56.40€
Option HC:      57.02€
Option Tempo:   82.67€
doudou@DGaming:~/go/edfbleu$
```

Exemple avec un export de consommation quotidienne (le programme prévient alors que les montants HC/HP et Tempo sont approximatifs, voir la note correspondante plus bas) :

```plain
doudou@DGaming:~/go/edfbleu$ ./edfbleu consommation_quotidienne.xlsx

/!\ daily export: the intraday consumption profile is unknown so each day has been spread uniformly over its half-hours. The base option is exact but the HC/HP and Tempo amounts are approximations.

PRM:            12345678901234
Start:          2025-01-01 00:00:00 +0100 CET
End:            2025-04-01 00:00:00 +0200 CEST
Stepping:       24h0m0s

* TOTAL
Consommation:   ~1416.75 kWh
Option base:    356.46€
Option HC:      352.71€
Option Tempo:   308.60€
doudou@DGaming:~/go/edfbleu$
```

## Notes

### Heures creuses

Différentes périodes d'heure creuses existent par commune (vous pouvez vérifiez [ici](https://www.enedis.fr/heures-creuses/standard)). Une commune peut d'ailleurs en avoir plusieurs différentes suivant les jours. Le calcul d'heures creuses de ce programme est donc une approximation, la période la plus répandue 23H30-7H30 a donc été sélectionnée pour les calculs. Gardez cependant en tête que votre commune a certainement des tranches horaires différentes et/ou dynamiques.

### Exports quotidiens

L'export de consommation quotidienne ne contient que la consommation totale de chaque jour : la répartition de votre consommation au cours de la journée (heures creuses/heures pleines et plages tempo) n'y figure pas. Le programme répartit donc chaque jour uniformément sur ses demi-heures pour effectuer les calculs : l'option base reste exacte, mais les montants des options heures creuses et tempo sont des approximations. Pour des résultats exacts, utilisez les exports de courbe de charge (demi-heure).

### Fusion de plusieurs exports

Le programme accepte plusieurs fichiers en entrée : leurs données sont fusionnées puis tarifées ensemble, ce qui permet de couvrir une longue période avec les exports hebdomadaires de courbe de charge. Tous les fichiers doivent porter sur le même PRM et le même type d'export (les exports quotidiens et les exports de courbe de charge ne peuvent pas être mélangés). Si deux exports couvrent des périodes qui se recouvrent, le programme s'arrête en signalant l'anomalie ; si des enregistrements manquent, un warning vous prévient que la consommation de ces périodes n'est pas comptée.

### Jours tempo

#### Tempo calendar fetch

At startup, the program tries to fetch the Tempo day colors from [api-couleur-tempo.fr](https://www.api-couleur-tempo.fr/), a free third-party API serving the RTE Tempo data (RTE remains the source of truth). If the API is unavailable, the program falls back to the Tempo calendar compiled into the binary (last update: 02/04/2024) and prints a warning: the fetch itself never makes the program fail. Days which have not been decided yet have no color and can not be priced: the program warns you when your data set goes beyond the last decided day.

#### Jours rouges et blancs

Dans le cas nominal, aucun warning n'apparaît. Vous n'en verrez un que dans deux situations : l'API est indisponible et votre export dépasse le calendrier compilé de secours (02/04/2024), ou votre export contient des dates dont la couleur n'a pas encore été décidée par RTE — dans ce dernier cas, ces jours ne peuvent pas être tarifés correctement. Vous aurez alors l'un des warnings suivants:

```plain
/!\ the data set contains values that are beyond the internal data this program has. Please update the code.
```

```plain
/!\ the data set contains values that are beyond the last tempo day decided so far (17/10/2026): the color of the next days does not exist yet so they can not be priced.
```

Si vous rencontrez l'un de ces warnings, n'hésitez pas à ouvrir une [issue](https://github.com/hekmon/edfbleu/issues) ou une [pull request](https://github.com/hekmon/edfbleu/pulls) afin d'en discuter.

#### Consommation sous tempo

Les valeurs que vous obtiendrez correspondent à votre consommation *passée* où vous ne décaliez pas forcément votre consommation. Le but de tempo étant de vous encourager à décaler votre consommation en journée rouge (prix du kWh prohibitif), votre prix tempo sera certainement bien moins élevé si vous décalez vos consommations pendant ces plages horaires. Plus d'information sur tempo [ici](https://particulier.edf.fr/fr/accueil/gestion-contrat/options/tempo/details.html).
