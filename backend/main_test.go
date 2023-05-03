package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
	"encoding/json"
    "bytes"
	"io"
	"strings"
)

func TestResetRegistry1(t *testing.T) {
    router := gin.Default()
    router.DELETE("/reset", resetRegistry)

    req, _ := http.NewRequest("DELETE", "/reset", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
    }
}

func TestValidInput(t *testing.T) {
    router := gin.Default()
    router.POST("/package", createPackage)
	router.GET("/package/:id", getPackageByID)
	router.PUT("/package/:id", updatePackageByID)
	router.DELETE("/package/:id", deletePackageByID)
	router.GET("/package/:id/rate", ratePackage)
	router.POST("/packages", searchPackages)
	router.POST("/package/byRegEx", searchByRegEx)

	// Test Content input 1
	zipData1 := CreatePackageRequest{
		Content: "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",
		JSProgram: "if (process.argv.length === 7) {\nconsole.log('Success')\nprocess.exit(0)\n} else {\nconsole.log('Failed')\nprocess.exit(1)\n}\n",
	}
	jsonZipData1, err := json.Marshal(zipData1)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ := http.NewRequest("POST", "/package", bytes.NewBuffer(jsonZipData1))

    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusBadRequest {
        t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
    }

	// Test Content input 2
	zipData2 := CreatePackageRequest{
		Content: "UEsDBAoAAAAAAK4NfkwAAAAAAAAAAAAAAAAPAAkAaXMtZXZlbi1tYXN0ZXIvVVQFAAEp+b1aUEsDBAoAAAAIAK4NfkzZFIrBnAAAAAUBAAAcAAkAaXMtZXZlbi1tYXN0ZXIvLmVkaXRvcmNvbmZpZ1VUBQABKfm9Wn2OQQ6CMBBF9z0F66ZK4sqNJzGkmbRTmaQMpJ0GlHB3gaA73c7/779JfS/VrZJUUKm7bhSxRxab5RlxDfIADhWyt32wkXi7xaBcCynjRhYJp+uXotdWuChJ1FlJQCvysGNLgvvSR0WcMYkNxBAt43gsH2/MWtczOCkQTaBJSsJscBrQCXoj2A0R1sGl1troc+eX5p8wQMw/jXv4BlBLAwQKAAAACACuDX5MKzlIZqkEAABVDgAAHQAJAGlzLWV2ZW4tbWFzdGVyLy5lc2xpbnRyYy5qc29uVVQFAAEp+b1anVdLayQ3EL77V0yGHJJggXcPOexlCYbAQkggPpoNVHfX9ChWS21JPY9s/N/zSeqnuicHY2w89fhUVfqqVPPtbrfbc9nQr0y+s+z2n3bfIIO0MVWnosDbju+TkC8tW9mw9qT+KP7m0v/Jzj+1lqnqLWH4dn8XcfVpgiusOTu2EBxIuRHP/bw8QJuKl5LGlEdaYdfKFMCZ8CtTdiGu7ABNJ1mTN/nBZ6krcx6EE67tc+5RqSzZOWNFS9IG+cfenyzyEa6lUuoa8ueP97tvyJIPxo4JwOzgY87h4+7ta+9cKFO+ZM57Ume6uv1oY6lk4fxV8WDxwRduH48hpcz5Cc6Kf5OaVweUpmlIVBQMBm/NJ4SytLgdf6rVrQR673lwipyfoWsHjxJ1F65rI8JQubKz6jp4NZ3yUqiQw+BbGS9QH/LS6MGstQYg/joasVEinjjh8mv4maqJCgndKTX61KzZUgzJk33v3R1JV4pFiQMKwi2ytcOZf/2AD//i19gfv9/vBg8wLfEyGKWjns7Sl8dHcuG4DxP4C19vxfVoVKzH4mIG4TJEoJyNrd6boeYz0msnNwi+uMco6R2h/uJ+57F/ls4tWdbzZtFGoF/oKma8WKpDOTmX4YKdwK+sdaYxurqlcP6WxluQxnLNl6Wq4qKr6/z0ihV7FifK5aAzsqndhjhF3HBTsN3S42LWYiXB9UCoSIaZjpvWo2RHwiTAPSX0zOREKpNcNtPni2fUTKOrTvkxF29JFGDpptwYxaRDeH5LP9517IBDp8vQt9MYgyEYovzRmq4+LhEOyiAeXaPWpWzyTALWZi6yQc242kheanR4QFOoWYrj/wKTGgiySpxoMygLaQcccT5Kz6GVsrpBamlFZEUFqzVroji7PLQui/gUZIpGXjh1L4P/uDZPRW4S52Yy2dR4uyFuMbgSq8LEHYqDDm/oshxE8EhcQXHGO3iYKRmPKoKUWqhjFoGOc2EpiLe5lpq4Qqzlll87aXmtOFtq21VzASZOkFxc+pwhUSTYYYBl6HhivFkmaTlRKbOMdNmsvWXsUHqTtI7VAUOoaVdwDrmyXmG5I2FBASSuUpah1pqalVEIotooLhTWcZq6mY8/SifSS7B6m6M6LDYqsjsrHvpdqtCvY+6zYnV45Q5L+ygCRaTP5WGNjDlFWqq0wiwsNKP4YD5bTfY6MRWIBJ9fYonn+96Cu50GbTE5C7UC7hxg0Z8z9vefws6wD69iGu4wR1hLWNwi1sHItWX2eM/nwy10dpoBzw/xiFAEiWHzD1cROu5iE3bYbuJmEkpRIPaX3rV/oiOGgY+VVVpPd/vPAah/0WH4afZx9zZBt1SFQo5zJqIud8HXzvhxGuxdXCtjIU4G07FvlsHYUiXnD6jjRt5YYYPq3Stm5NjA0yn6rWPmlsOgTw/T9u6bHKTOHq9to4O8CNPOeyhpusDLXtNnFpau8ftS5M8gyRjat23YovvN8GFKqx/I9iWNuef+q064kOFLz3344gSq+Ok/UUkX6Q7JT4B/oTr+/134c7//Oh0OEgvpNM0HVHoH/bVlM+/hMG2FlIdxySc9reBXU1FWuvA16u7t7j9QSwMECgAAAAgArg1+TCmuqMFQAAAAfwAAAB0ACQBpcy1ldmVuLW1hc3Rlci8uZ2l0YXR0cmlidXRlc1VUBQABKfm9WlNWcM1Lyy9KTlUIzcusUMhLLc/JzEst5tJSKEmtKFFIzc+xzUnj4lJWSMrMSyzKBMnoJWZCeJVAdkFxCoKTVZCO4KRnpiEpy0tHVpYK4wEAUEsDBAoAAAAIAK4Nfkzwb3f+ugAAAA8BAAAZAAkAaXMtZXZlbi1tYXN0ZXIvLmdpdGlnbm9yZVVUBQABKfm9WjWOQY7DMAhF95wCqbtqkl6i+y46e8uxqWvVhgg7rXL7wVFnAR99nj6c0JeP3xvmxKKEj1yowXm+3t29m2Fj25aSK01ngBN2ah2Viu8Uf1AUY1YKRmZqmIhJxwaX/SAbjH7xoW++wFeCvI1KBDPv4QjltQJLJFclbuO+GVOkZUtzkTSI3SvDaGaE1zFNpCr6D9TcArj0dKslN3BR/aN/pcEiH1IXpK7CxGa8iaOofVdX6Fa/t+ttrhH+AFBLAwQKAAAACACuDX5MOlHB208AAAB8AAAAGgAJAGlzLWV2ZW4tbWFzdGVyLy50cmF2aXMueW1sVVQFAAEp+b1aKy5NybdSSEvMKU7lyi+24lJQ0FXIycwrrQCz8osruHIS89JLE9NTrRTy8lNS47OKuaA0RDGIA2aom6tDaDMobQqlTaC0gZ6hEYJpoM4FAFBLAwQKAAAACACuDX5MJs/1CF0AAACUAAAAFwAJAGlzLWV2ZW4tbWFzdGVyLy52ZXJiLm1kVVQFAAEp+b1aU1ZWCC1OTE/l4kpISMgq5ipLLFLILHYtS81TsFUoSi0szSxK1VCvVrVVyEvMTVVQrVXXtObigqjQMACy9fVt7RRKikpTYYLqhuow4bTEnGK4uBFWxcZoioGu4AIAUEsDBAoAAAAIAK4Nfky1q+EYgwIAAEAEAAAWAAkAaXMtZXZlbi1tYXN0ZXIvTElDRU5TRVVUBQABKfm9Wl1S3W+bMBB/919xylMroXabNE3amwNO441gZJxmeSTgBG8ER9hZ1P9+dyRt1UkIdF+/jztMZ2ElDeSusUOwcIfBPWOpP72M7tBFuGvu4cunz18Ten9L4IcfoGq63g1/7BgZK+14dCE4TLsAnR3t7gUOYz1E2yawH60Fv4emq8eDTSB6qIcXONkx4IDfxdoNbjhADQ0yMuyMHcIEv4+XerTY3EIdgm9cjXjQ+uZ8tEOsI/HtXW8D3EW0MKtuE7P7iaS1dc/cAFR7LcHFxc6fI4w2xNE1hJGAG5r+3JKG13Lvju7GQOPTGgJD0HNAB6QzgaNv3Z6+drJ1Ou96F7oEWkfQu3PEZKDktNWEfDz6EYLte4YIDnVPXt/VTT0k/UQLjbcVBcpcOn/86MQFtj+PA1Laaab1uLKJ8bdtImWofe/73l/IWuOH1pGj8J0xg6V65//aycv1yoOPKPUqgQ5wer/qrRS6uu9hZ28LQ143MEq92hmJPkQ8vKt7OPlx4vvf5gPyLwVUamE2XAuQFZRaPctMZDDjFcazBDbSLNXaAHZoXpgtqAXwYgs/ZZElIH6VWlQVKM3kqsylwJws0nydyeIJ5jhXKPyhJf7JCGoUEOENSoqKwFZCp0sM+Vzm0mwTtpCmIMyF0sCh5NrIdJ1zDeVal6oSSJ8hbCGLhUYWsRKFeUBWzIF4xgCqJc9zomJ8jeo16YNUlVstn5YGlirPBCbnApXxeS6uVGgqzblcJZDxFX8S05RCFM2o7aoONktBKeLj+KRGqoJspKowGsMEXWrzNrqRlUiAa1nRQhZarRJG68QJNYHgXCGuKLRq+HARbKF4XYk3QMgEzxELz1N8ON8D+wdQSwMECgAAAAgArg1+TKqGBgJyAwAAUQgAABgACQBpcy1ldmVuLW1hc3Rlci9SRUFETUUubWRVVAUAASn5vVqlVNtu4zYQfedXTOEitoRYSlJgC2yRLdIgBVIkiyBJHwojSChpLDFLkS4vUl30h/ob/bIOKSs32O1uCxgWL3PmzOUMJyDsHDtUsPhq8fHqEjo0Vmh1N2ucW9n3eS7aOrONQFnZTOhcrdq8yzegzHb199atJR4vJXfJM6rv+4xMH21W6jZf8fITr3GEJSNZq5Vr5Boq3SupeWX/kbZq/5134NSm3snptOPycxnd/2W8EMr/Bj94ISu4cdz53YTO8E7Y/FErWzZSqE9odtHvSV6gPL6NiBexDC7mpYjxbPWUMPYBrtF5o8AZjyCW4BqEWgQRKN8WaEgTEFkZm0zgXFmqmGRss4BeuAYWlPeudifvGXt4eLAN+xroFMQGOJ9b3uGouGASCX62VLWIeLSs44H+LARzDAZ/9cLgbLqBTJPvGBtuZwe0zvPjDzGL8XB6OB2Pl1zap/OjrcbfvDEeAzoptHdhNYFTUqgRhXdC1YxdecoiBIXWWeCqAkrM0MogcNnztYUeJZUAM/hRGyh8PZgtkVpPRiN2HxYridwilIauCK0oa+vxbpZlOf3ixuYK+yR7E4k2lrE/IE1PddsKZ9MUht3TfTwhk/l8Dpt/2r2j9eKVJp77V1NLfRGb98oiicjDgFR6qU2rdEk8qLZj39gEdIw96p8qSGNXUvD3s9uGJLYy+hFLN7VUFl5RzdoqKK9GhYZqUkGxhgW9R8VWrnChLcbvfMTMB0/JPmzKW2k1dYCVcFHlwzVUJKrSyXUGJ2oNZcNVjZZehpc2rbcOCvryiiSr4tUiC2wUJ7Vps0pGe4ftikYTs+SesVv9lMYLn/tg/OBoqaXUfagIpdKSQnZNTA0v85xU2MG2hGFvL56PEp7AtVcq+HdBboyN2yBGg53ASO5VqEsUMxWeQx3ECKTjUIsaHc1FK6TgRvxO7Yhzz0GKwnCzjq5IfnBydZ7BL9pDGUS8CbzCFaoKVSlwmICYeWSKbr6kCJRb2Ab0c34n3jXaMJamP2kFN0+KTVM6g8UglPxL5R6gjgJ0aHZhN9fbwDGwC1GioseEnerV2oi6cfDXn3B0cPgtDf3rWD8joIxdY1RyRc2q6GmOMrw8vx157mYX56dnH2/OwjuRhuzv43AthUTq5LZxeiue/zJe3UH2LjvYB8rnkuRyRMmFFLP7vwFQSwMECgAAAAgArg1+TI+fEBvJAAAA/gAAABcACQBpcy1ldmVuLW1hc3Rlci9pbmRleC5qc1VUBQABKfm9Wj2OwU7DMAyG73kK79R2Gg0gISQKXBAHEAgJeIEtMYuhS4rtTCDEu+MixMWH399n/365cLAEkgPcY4bzpDrJmfdb0lQ3fSg7/1qyhDRSfkNW/0demjWLV2X6ZNomhTZ0cHx4dLKa5+kKbkuGp3+vn+FHHHEtGKHmiAyaEO5vnuGOAmbBGfHONVUQRJmCNoNz+zVbu4cY4QIY3ysxto2VKDE2ne13JdYRe/yYCqsY9FJzULLnJNdWtKUOvhyYq5UzLH5vWTi478H9AFBLAwQKAAAACACuDX5MvtRn66gBAADOAwAAGwAJAGlzLWV2ZW4tbWFzdGVyL3BhY2thZ2UuanNvblVUBQABKfm9WpVTzY7TMBC+9ymsnECiabtwWglOe1kkLsANLZJrT5LZOuNoPC5Uq747ttOkKUIgTvGMv5+Z8eRlpVRFuofqXlUY1nAEqt7kpIVgGAdBT/nuM0hkUsIRFDZKOlAtJrCi2O+BFQaVufVIPgKHC3FXb+vtmO18D4Nui1cnMoT7zaZF6eK+Nr7fPHsKpnNIB2DZ3BSjo3SeM++jJ/VlRqlX/9J5PQowDD6geD5lkb847WMbEuQlnVMU2f1PsekbIoQqkc9FzaEBCqXhT49fR4cGHWSLb6MFkoWf9XMhPRVAr7FMbr4pWaAWCRa1kbdF+MP7bb3LM55dx4dbQAWCZGjvTaevOAsDJA8yuNRNvXhrM/x7Eq7vlvjjwx8pbXTDuvHca1n3V+rurpSuJuOcf1u/W0oe4PTDs13MQzPr08QzPpJMwfxI4xAFeIqQBNpFGGZfLd10TosKjGYKc4+XYxBGaq8PkLZ3vxieNylotAtwwTt98rEM1EKjo5srFB0O11bK2mmb/q0SPl1Ag4vpJW9gv83vBp5WTOZiimSTty7z889Y0uc8z9V59QtQSwMECgAAAAgArg1+TOjl7XFkAQAAdwMAABYACQBpcy1ldmVuLW1hc3Rlci90ZXN0LmpzVVQFAAEp+b1apZJNT8JAEIbv/RVjYrKFQLeFeAH1Yjxo9KRHL+126K7ALu4HaAz/3V1aCB/WkHjo5X3nnXlmp7R7EUEXhOnjEiVcc2sXZkRpJSx3RcLUnL4raRifCTlFbWlTeetTIXinFl9aVNxCzDowSLMreFQSXnaBJFQ9CYbSYAlOlqjBcoTnh1eY1XIooVFEnEEwVgtmyTiKNH44oTEmc8V4TjrjaJlryI3xTeEGdnatbH1h7sMee35CgxeVaJgWhRfqEtKDiZPMCiXjDnxHAMLGxHDlZqUPW6clWO0QxGTDK928wNAfwvqj0zg0bHHdP047fuyefNHo2ZHeyIOW8mGtr8MOB4wrpaew8mfavJmszBlMJCUtY0h27GyNQWtkSNrYLNdqBbkE1Fpp8D9EkZeQ6z8hk03KxCc2NFeNG451Dyh+LpBZA3lzmLfkkp4NI5XsC2mx8gf9NxTJkgxT8hubhGbKAV34fgBQSwECAAAKAAAAAACuDX5MAAAAAAAAAAAAAAAADwAJAAAAAAAAABAAAAAAAAAAaXMtZXZlbi1tYXN0ZXIvVVQFAAEp+b1aUEsBAgAACgAAAAgArg1+TNkUisGcAAAABQEAABwACQAAAAAAAQAAAAAANgAAAGlzLWV2ZW4tbWFzdGVyLy5lZGl0b3Jjb25maWdVVAUAASn5vVpQSwECAAAKAAAACACuDX5MKzlIZqkEAABVDgAAHQAJAAAAAAABAAAAAAAVAQAAaXMtZXZlbi1tYXN0ZXIvLmVzbGludHJjLmpzb25VVAUAASn5vVpQSwECAAAKAAAACACuDX5MKa6owVAAAAB/AAAAHQAJAAAAAAABAAAAAAACBgAAaXMtZXZlbi1tYXN0ZXIvLmdpdGF0dHJpYnV0ZXNVVAUAASn5vVpQSwECAAAKAAAACACuDX5M8G93/roAAAAPAQAAGQAJAAAAAAABAAAAAACWBgAAaXMtZXZlbi1tYXN0ZXIvLmdpdGlnbm9yZVVUBQABKfm9WlBLAQIAAAoAAAAIAK4Nfkw6UcHbTwAAAHwAAAAaAAkAAAAAAAEAAAAAAJAHAABpcy1ldmVuLW1hc3Rlci8udHJhdmlzLnltbFVUBQABKfm9WlBLAQIAAAoAAAAIAK4Nfkwmz/UIXQAAAJQAAAAXAAkAAAAAAAEAAAAAACAIAABpcy1ldmVuLW1hc3Rlci8udmVyYi5tZFVUBQABKfm9WlBLAQIAAAoAAAAIAK4Nfky1q+EYgwIAAEAEAAAWAAkAAAAAAAEAAAAAALsIAABpcy1ldmVuLW1hc3Rlci9MSUNFTlNFVVQFAAEp+b1aUEsBAgAACgAAAAgArg1+TKqGBgJyAwAAUQgAABgACQAAAAAAAQAAAAAAewsAAGlzLWV2ZW4tbWFzdGVyL1JFQURNRS5tZFVUBQABKfm9WlBLAQIAAAoAAAAIAK4NfkyPnxAbyQAAAP4AAAAXAAkAAAAAAAEAAAAAACwPAABpcy1ldmVuLW1hc3Rlci9pbmRleC5qc1VUBQABKfm9WlBLAQIAAAoAAAAIAK4Nfky+1GfrqAEAAM4DAAAbAAkAAAAAAAEAAAAAADMQAABpcy1ldmVuLW1hc3Rlci9wYWNrYWdlLmpzb25VVAUAASn5vVpQSwECAAAKAAAACACuDX5M6OXtcWQBAAB3AwAAFgAJAAAAAAABAAAAAAAdEgAAaXMtZXZlbi1tYXN0ZXIvdGVzdC5qc1VUBQABKfm9WlBLBQYAAAAADAAMALkDAAC+EwAAKAA1ODVmODAwMmJiMTZmN2JlYzcyM2E0NzM0OWI2N2RmNDUxZjFiMjVk",
		JSProgram: "if (process.argv.length === 7) {\nconsole.log('Success')\nprocess.exit(0)\n} else {\nconsole.log('Failed')\nprocess.exit(1)\n}\n",
	}
	jsonZipData2, err := json.Marshal(zipData2)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ = http.NewRequest("POST", "/package", bytes.NewBuffer(jsonZipData2))

    resp = httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusCreated {
        t.Errorf("Expected status code %d, but got %d", http.StatusCreated, resp.Code)
    } else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseData UpdatePackageRequest
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		// Test download
		req, _ = http.NewRequest("GET", ("/package/" + responseData.Metadata.ID), nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}

		// Test rate
		req, _ = http.NewRequest("GET", ("/package/" + responseData.Metadata.ID + "/rate"), nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}

		// Test Update
		req, _ = http.NewRequest("PUT", ("/package/" + responseData.Metadata.ID), bytes.NewBuffer(body))
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}

		// Test delete
		req, _ = http.NewRequest("DELETE", ("/package/" + responseData.Metadata.ID), nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}
	}

	// Test Content URL input 1
	urlData1 := CreatePackageRequest{
		JSProgram: "if (process.argv.length === 7) {\nconsole.log('Success')\nprocess.exit(0)\n} else {\nconsole.log('Failed')\nprocess.exit(1)\n}\n",
		URL: "https://github.com/microsoft/restler-fuzzer",
	}
	jsonUrlData1, err := json.Marshal(urlData1)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ = http.NewRequest("POST", "/package", bytes.NewBuffer(jsonUrlData1))

    resp = httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusBadRequest {
        t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
    }

	// Test URL input 2
	urlData2 := CreatePackageRequest{
		JSProgram: "if (process.argv.length === 7) {\nconsole.log('Success')\nprocess.exit(0)\n} else {\nconsole.log('Failed')\nprocess.exit(1)\n}\n",
		URL: "https://github.com/jashkenas/underscore",
	}
	jsonUrlData2, err := json.Marshal(urlData2)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ = http.NewRequest("POST", "/package", bytes.NewBuffer(jsonUrlData2))

    resp = httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusCreated {
        t.Errorf("Expected status code %d, but got %d", http.StatusCreated, resp.Code)
    } else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseData UpdatePackageRequest
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}
		responseData.Data.Content = ""
		responseData.Data.URL = "https://github.com/jashkenas/underscore"
		newData, err := json.Marshal(responseData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON data: %v", err)
		}

		// Test Update
		req, _ = http.NewRequest("PUT", ("/package/" + responseData.Metadata.ID), bytes.NewBuffer(newData))
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}
	}

	req, _ = http.NewRequest("POST", "/package", bytes.NewBuffer(jsonUrlData2))

    resp = httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusConflict {
        t.Errorf("Expected status code %d, but got %d", http.StatusConflict, resp.Code)
    }

	// Test fetching all packages
	packQuery := `[{"Name": "*"}]`
    req, _ = http.NewRequest("POST", "/packages?offset=1", strings.NewReader(packQuery))
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
	}

	packQuery = `[{"Name": "underscore", "Version": "1.13.6"}]`
    req, _ = http.NewRequest("POST", "/packages?offset=1", strings.NewReader(packQuery))
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
	}

	// Test search by regex
	packRegex := PackageByRegex{
		RegEx: "underscore",
	}
	jsonPackRegex, err := json.Marshal(packRegex)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ = http.NewRequest("POST", "/package/byRegEx", bytes.NewBuffer(jsonPackRegex))
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
	}
}

func TestResetRegistry2(t *testing.T) {
    router := gin.Default()
    router.DELETE("/reset", resetRegistry)

    req, _ := http.NewRequest("DELETE", "/reset", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
    }
}

func TestInvalidInput(t *testing.T) {
    router := gin.Default()
	router.GET("/package/:id", getPackageByID)
	router.PUT("/package/:id", updatePackageByID)
	router.DELETE("/package/:id", deletePackageByID)
	router.GET("/package/:id/rate", ratePackage)
	router.POST("/packages", searchPackages)
	router.POST("/package/byRegEx", searchByRegEx)
	router.POST("/package", createPackage)

	// Test upload
	req, _ := http.NewRequest("POST", "/package", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
	}

	// Test download
	req, _ = http.NewRequest("GET", "/package/invalid", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, resp.Code)
	}

	// Test rate
	req, _ = http.NewRequest("GET", "/package/invalid/rate", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, resp.Code)
	}

	// Test Update
	req, _ = http.NewRequest("PUT", "/package/invalid", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
	}

	zipData := CreatePackageRequest{
		Content: "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",
		JSProgram: "if (process.argv.length === 7) {\nconsole.log('Success')\nprocess.exit(0)\n} else {\nconsole.log('Failed')\nprocess.exit(1)\n}\n",
	}
	jsonZipData, err := json.Marshal(zipData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

	req, _ = http.NewRequest("POST", "/package", bytes.NewBuffer(jsonZipData))

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, but got %d", http.StatusCreated, resp.Code)
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseData UpdatePackageRequest
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		// Test Update
		req, _ = http.NewRequest("PUT", "/package/invalid", bytes.NewBuffer(body))
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, resp.Code)
		}

		// Test delete
		req, _ = http.NewRequest("DELETE", ("/package/" + responseData.Metadata.ID), nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
		}
	}

	// Test delete
	req, _ = http.NewRequest("DELETE", "/package/invalid", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, resp.Code)
	}

	// Test search by regex
	packRegex := PackageByRegex{
		RegEx: "underscore",
	}
	jsonPackRegex, err := json.Marshal(packRegex)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

    req, _ = http.NewRequest("POST", "/package/byRegEx", bytes.NewBuffer(jsonPackRegex))
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, resp.Code)
	}

	req, _ = http.NewRequest("POST", "/package/byRegEx", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
	}

	// Test fetching all packages
    req, _ = http.NewRequest("POST", "/packages?offset=1", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestResetRegistry3(t *testing.T) {
    router := gin.Default()
    router.DELETE("/reset", resetRegistry)

    req, _ := http.NewRequest("DELETE", "/reset", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.Code)
    }
}