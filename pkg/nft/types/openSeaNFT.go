package types

type OpenSeaNFTReq struct {
	Id        string      `json:"id"`
	Variables interface{} `json:"variables"`
	Query     string      `json:"query"`
}

type OpenSeaAsset struct {
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code string `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
	Data struct {
		Nft struct {
			IsItemType string `json:"__isItemType"`
			RelayID    string `json:"relayId"`
			Typename   string `json:"__typename"`
			Chain      struct {
				Identifier string `json:"identifier"`
			} `json:"chain"`
			Decimals             interface{} `json:"decimals"`
			FavoritesCount       int         `json:"favoritesCount"`
			IsDelisted           bool        `json:"isDelisted"`
			IsFrozen             bool        `json:"isFrozen"`
			HasUnlockableContent bool        `json:"hasUnlockableContent"`
			IsFavorite           bool        `json:"isFavorite"`
			TokenID              string      `json:"tokenId"`
			Name                 string      `json:"name"`
			AuthenticityMetadata interface{} `json:"authenticityMetadata"`
			ImageURL             string      `json:"imageUrl"`
			Creator              struct {
				Address string `json:"address"`
				ID      string `json:"id"`
				User    struct {
					PublicUsername string `json:"publicUsername"`
					ID             string `json:"id"`
				} `json:"user"`
				DisplayName   string `json:"displayName"`
				Config        string `json:"config"`
				IsCompromised bool   `json:"isCompromised"`
				ImageURL      string `json:"imageUrl"`
			} `json:"creator"`
			AssetContract struct {
				Address           string      `json:"address"`
				Chain             string      `json:"chain"`
				BlockExplorerLink string      `json:"blockExplorerLink"`
				ID                string      `json:"id"`
				OpenseaVersion    interface{} `json:"openseaVersion"`
				TokenStandard     string      `json:"tokenStandard"`
			} `json:"assetContract"`
			AnimationURL    string      `json:"animationUrl"`
			BackgroundColor interface{} `json:"backgroundColor"`
			Collection      struct {
				Description string `json:"description"`
				DisplayData struct {
					CardDisplayStyle string `json:"cardDisplayStyle"`
				} `json:"displayData"`
				Category struct {
					Slug string `json:"slug"`
				} `json:"category"`
				Hidden             bool          `json:"hidden"`
				ImageURL           string        `json:"imageUrl"`
				Name               string        `json:"name"`
				Slug               string        `json:"slug"`
				VerificationStatus string        `json:"verificationStatus"`
				IsCategory         bool          `json:"isCategory"`
				NumericTraits      []interface{} `json:"numericTraits"`
				StatsV2            struct {
					TotalSupply float64 `json:"totalSupply"`
					FloorPrice  struct {
						Unit   string `json:"unit"`
						Symbol string `json:"symbol"`
						Eth    string `json:"eth"`
					} `json:"floorPrice"`
				} `json:"statsV2"`
				RelayID                  string      `json:"relayId"`
				DiscordURL               string      `json:"discordUrl"`
				ExternalURL              string      `json:"externalUrl"`
				InstagramUsername        interface{} `json:"instagramUsername"`
				MediumUsername           interface{} `json:"mediumUsername"`
				TelegramURL              interface{} `json:"telegramUrl"`
				TwitterUsername          interface{} `json:"twitterUsername"`
				ConnectedTwitterUsername string      `json:"connectedTwitterUsername"`
				AssetContracts           struct {
					Edges []struct {
						Node struct {
							BlockExplorerLink string `json:"blockExplorerLink"`
							ChainData         struct {
								BlockExplorer struct {
									Name       string `json:"name"`
									Identifier string `json:"identifier"`
								} `json:"blockExplorer"`
							} `json:"chainData"`
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"assetContracts"`
				EnabledRarities []interface{} `json:"enabledRarities"`
				ID              string        `json:"id"`
				PaymentAssets   []struct {
					RelayID string `json:"relayId"`
					Symbol  string `json:"symbol"`
					Chain   struct {
						Identifier string `json:"identifier"`
					} `json:"chain"`
					Asset struct {
						UsdSpotPrice    float64 `json:"usdSpotPrice"`
						Decimals        int     `json:"decimals"`
						ID              string  `json:"id"`
						RelayID         string  `json:"relayId"`
						DisplayImageURL string  `json:"displayImageUrl"`
					} `json:"asset"`
					IsNative bool   `json:"isNative"`
					ID       string `json:"id"`
				} `json:"paymentAssets"`
				Logo string `json:"logo"`
			} `json:"collection"`
			Description          string `json:"description"`
			NumVisitors          int    `json:"numVisitors"`
			IsListable           bool   `json:"isListable"`
			IsReportedSuspicious bool   `json:"isReportedSuspicious"`
			IsUnderReview        bool   `json:"isUnderReview"`
			IsCompromised        bool   `json:"isCompromised"`
			IsBiddingEnabled     struct {
				Value  bool        `json:"value"`
				Reason interface{} `json:"reason"`
			} `json:"isBiddingEnabled"`
			Traits struct {
				Edges []struct {
					Node struct {
						RelayID     string      `json:"relayId"`
						DisplayType interface{} `json:"displayType"`
						FloatValue  interface{} `json:"floatValue"`
						IntValue    interface{} `json:"intValue"`
						TraitType   string      `json:"traitType"`
						Value       string      `json:"value"`
						TraitCount  int         `json:"traitCount"`
						MaxValue    interface{} `json:"maxValue"`
						ID          string      `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"traits"`
			DefaultRarityData           interface{} `json:"defaultRarityData"`
			DisplayImageURL             string      `json:"displayImageUrl"`
			ExternalLink                interface{} `json:"externalLink"`
			ImageStorageURL             string      `json:"imageStorageUrl"`
			VerificationStatus          string      `json:"verificationStatus"`
			MetadataStatus              string      `json:"metadataStatus"`
			FrozenAt                    interface{} `json:"frozenAt"`
			TokenMetadata               string      `json:"tokenMetadata"`
			LastUpdatedAt               string      `json:"lastUpdatedAt"`
			OpenseaSellerFeeBasisPoints int         `json:"openseaSellerFeeBasisPoints"`
			TotalCreatorFee             int         `json:"totalCreatorFee"`
			OwnedQuantity               interface{} `json:"ownedQuantity"`
			AssetOwners                 struct {
				Edges []struct {
					Node struct {
						Quantity string `json:"quantity"`
						Owner    struct {
							Address       string      `json:"address"`
							Config        interface{} `json:"config"`
							IsCompromised bool        `json:"isCompromised"`
							User          struct {
								PublicUsername string `json:"publicUsername"`
								ID             string `json:"id"`
							} `json:"user"`
							DisplayName string `json:"displayName"`
							ImageURL    string `json:"imageUrl"`
							ID          string `json:"id"`
						} `json:"owner"`
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				Count int `json:"count"`
			} `json:"assetOwners"`
			TotalQuantity string `json:"totalQuantity"`
			LargestOwner  struct {
				Owner struct {
					Address string `json:"address"`
					ID      string `json:"id"`
				} `json:"owner"`
				ID string `json:"id"`
			} `json:"largestOwner"`
			IsCurrentlyFungible bool   `json:"isCurrentlyFungible"`
			DisplayName         string `json:"displayName"`
			ID                  string `json:"id"`
		} `json:"nft"`
		TradeSummary struct {
			BestAsk interface{} `json:"bestAsk"`
			BestBid struct {
				PriceType struct {
					Unit string `json:"unit"`
					Usd  string `json:"usd"`
				} `json:"priceType"`
				PerUnitPriceType struct {
					Unit   string `json:"unit"`
					Usd    string `json:"usd"`
					Symbol string `json:"symbol"`
				} `json:"perUnitPriceType"`
				DutchAuctionFinalPriceType interface{} `json:"dutchAuctionFinalPriceType"`
				OpenedAt                   string      `json:"openedAt"`
				ClosedAt                   string      `json:"closedAt"`
				Payment                    struct {
					Symbol  string `json:"symbol"`
					ID      string `json:"id"`
					RelayID string `json:"relayId"`
				} `json:"payment"`
				ID      string `json:"id"`
				RelayID string `json:"relayId"`
			} `json:"bestBid"`
		} `json:"tradeSummary"`
		AcceptHighestOffer struct {
			BestBid struct {
				RelayID string `json:"relayId"`
				ID      string `json:"id"`
				Item    struct {
					Typename   string `json:"__typename"`
					IsItemType string `json:"__isItemType"`
					RelayID    string `json:"relayId"`
					Chain      struct {
						Identifier string `json:"identifier"`
					} `json:"chain"`
					TokenID       string `json:"tokenId"`
					AssetContract struct {
						Address string `json:"address"`
						ID      string `json:"id"`
					} `json:"assetContract"`
					IsNode     string `json:"__isNode"`
					ID         string `json:"id"`
					Collection struct {
						StatsV2 struct {
							FloorPrice struct {
								Eth string `json:"eth"`
							} `json:"floorPrice"`
						} `json:"statsV2"`
						ID string `json:"id"`
					} `json:"collection"`
					OwnedQuantity interface{} `json:"ownedQuantity"`
				} `json:"item"`
				PerUnitPriceType struct {
					Unit   string `json:"unit"`
					Symbol string `json:"symbol"`
					Eth    string `json:"eth"`
				} `json:"perUnitPriceType"`
				Side  string `json:"side"`
				Maker struct {
					Address string `json:"address"`
					ID      string `json:"id"`
				} `json:"maker"`
			} `json:"bestBid"`
		} `json:"acceptHighestOffer"`
		EventActivity struct {
			Edges []struct {
				Node struct {
					RelayID string `json:"relayId"`
					ID      string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"eventActivity"`
		TradeLimits struct {
			MinimumBidUsdPrice string `json:"minimumBidUsdPrice"`
		} `json:"tradeLimits"`
	} `json:"data"`
}
