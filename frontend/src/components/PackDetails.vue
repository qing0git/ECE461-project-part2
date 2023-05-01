<template>
  <div id="body-wrapper">
    <div v-if="loaded">
      <h4>Package Details</h4>
      <ul class="multi-row-list">
        <li>
          <div class="grid-container list-row list-title">
            <div class="grid-item first-column shaded-row">Name</div>
            <div class="grid-item shaded-row"> {{ packmetadata.Name }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column">Version</div>
            <div class="grid-item"> {{ packmetadata.Version }} </div>
          </div>
        </li>
        <br>
        <li v-if="showRate">
          <div class="grid-container list-row list-title">
            <div class="grid-item first-column shaded-row">Metric</div>
            <div class="grid-item shaded-row">Score</div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column">BusFactor</div>
            <div class="grid-item"> {{ rate.BusFactor.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column shaded-row">Correctness</div>
            <div class="grid-item shaded-row"> {{ rate.Correctness.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column">RampUp</div>
            <div class="grid-item"> {{ rate.RampUp.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column shaded-row">ResponsiveMaintainer</div>
            <div class="grid-item shaded-row"> {{ rate.ResponsiveMaintainer.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column">LicenseScore</div>
            <div class="grid-item"> {{ rate.LicenseScore.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column shaded-row">GoodPinningPractice</div>
            <div class="grid-item shaded-row"> {{ rate.GoodPinningPractice.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column">PullRequest</div>
            <div class="grid-item"> {{ rate.PullRequest.toFixed(1) }} </div>
          </div>
          <div class="grid-container list-row">
            <div class="grid-item first-column shaded-row">NetScore</div>
            <div class="grid-item shaded-row"> {{ rate.NetScore.toFixed(1) }} </div>
          </div>
        </li>
      </ul>
      <br>
      <button class="btn btn-ada" type="button" @click="downloadPack">Download</button>&nbsp;&nbsp;
      <button class="btn btn-ada" type="button" @click="showUpdate=!showUpdate">Update</button>&nbsp;&nbsp;
      <button class="btn btn-ada" type="button" @click="ratePack">Rate</button>
    </div>
    <div v-else>
      <h4>Processing</h4>
      <span class="loader"></span>
    </div>

    <div v-if="showUpdate">
      <br>
      <div id="body-wrapper">
        <div v-if="!submitted">
          <fieldset class="form-group">
            <div>
              <label style="padding-right: 10px;">Resource Type</label>
              <input
                type="radio"
                id="urlResource"
                name="resourceType"
                value="url"
                v-model="resourceType"
                aria-label="URL"
                placeholder="Upload with URL"
              />
              <label style="padding-left: 5px; padding-right: 10px" for="urlResource">URL</label>
              <input
                type="radio"
                id="zipResource"
                name="resourceType"
                value="zip"
                v-model="resourceType"
                label="Upload Zip File"
              />
              <label style="padding-left: 5px" for="zipResource">Zip File</label>
            </div>
          </fieldset>
          <br>
          <div class="form-group">
            <input
              type="text"
              class="form-control"
              id="jsProgram"
              required
              v-model="packdata.JSProgram"
              name="jsProgram"
              label="Input JS program code"
              placeholder="JS Program Code"
            />          
          </div>
          <br>
          <div class="input-group">
            <input
              v-if="resourceType === 'url'"
              type="text"
              class="form-control"
              id="url"
              required
              v-model="packdata.URL"
              name="url"
              label="URL to resource"
              placeholder="URL to resource"
            />
            <input
              v-if="resourceType === 'zip'"
              type="file"
              class="form-control"
              id="zipFile"
              required
              ref="zipInput"
              accept=".zip,application/zip,application/x-zip,application/x-zip-compressed,application/octet-stream"
              name="zipFile"
              label="Choose your zipfile"
            />
            <button @click="savePack" class="btn btn-ada" aria-label="Submit the form">
              Submit
            </button>
          </div>
        </div>
        <div v-else>
          <h4>Processing</h4>
          <span class="loader"></span>
        </div>
      </div>
    </div>
  </div>
  <footer class="fixed-bottom">
    <button @click="deletePack()" class="btn btn-danger" aria-label="Delete Package">Delete</button>
  </footer>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import PackService from "@/services/PackService";
import { 
  PackageMetadata,
  PackageData,
  PackageRating,
} from "@/types/Pack";
import JSZip from "jszip";
import { saveAs } from "file-saver";

export default defineComponent({
  name: "SampleTutorial",
  data() {
    return {
      packmetadata: {} as PackageMetadata,
      packdata: {} as PackageData,
      rate: {} as PackageRating,
      resourceType: "url",
      showUpdate: false,
      showRate: false,
      rated: false,
      loaded: false,
      submitted: false,
    };
  },
  methods: {
    async savePack() {
      let data: any = {
        metadata: {
          Name: this.packmetadata.Name,
          Version: this.packmetadata.Version,
          ID: this.packmetadata.ID,
        },
        data: {
          JSProgram: this.packdata.JSProgram,
        },
      };

      if (this.resourceType === "url") {
        data.data.URL = this.packdata.URL;
      } else {
        const zipInput = this.$refs.zipInput as HTMLInputElement;
        const file = zipInput.files?.[0];

        if (file) {
          const fileReader = new FileReader();
          await new Promise((resolve, reject) => {
            fileReader.onload = () => {
              // Get the base64 string of the uploaded zip file
              const base64WithPrefix = fileReader.result as string;
              // Remove the prefix
              data.data.Content = base64WithPrefix.split(",")[1];
              resolve(null);
            };
            fileReader.onerror = () => {
              fileReader.abort();
              reject(new Error("Error reading file."));
            };
            // Read the file as a Data URL, which will be a base64-encoded string
            fileReader.readAsDataURL(file);
          });
        }
      }
      this.submitted = true;

      PackService.updateByID(this.$route.params.id, data)
        .then((response) => {
          console.log(response.data);
          alert("Done");
          this.showUpdate = false;
        })
        .catch((error) => {
          console.log(error);
          alert("An error occured");
          this.submitted = false;
        });
    },

    async downloadPack() {
      try {
        const base64String = this.packdata.Content;
        if (!base64String) {
          console.error('No base64 string found in the JSON object');
          return;
        }

        // Convert the base64 string to a Uint8Array
        const binaryString = atob(base64String);
        const len = binaryString.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        const binaryData = bytes.buffer;

        // Load the zip content using jszip
        const zip = new JSZip();
        const loadedZip = await zip.loadAsync(binaryData);

        // Generate a new zip Blob for downloading
        const downloadBlob = await loadedZip.generateAsync({ type: 'blob' });

        // Create a hidden anchor element to trigger the download
        const link = document.createElement('a');
        link.href = URL.createObjectURL(downloadBlob);
        link.download = 'downloaded-pack.zip';
        link.style.display = 'none';
        document.body.appendChild(link);
        link.click();

        // Clean up
        setTimeout(() => {
          URL.revokeObjectURL(link.href);
          document.body.removeChild(link);
        }, 100);
      } catch (error) {
        console.error('Error while downloading the pack:', error);
      }
    },

    ratePack() {
      if (this.rated) {
        this.showRate = !this.showRate;
      } else {
      PackService.ratePack(this.$route.params.id)
        .then((response) => {
          this.rate = response.data;
          this.showRate = true;
          this.rated = true;
        })
        .catch((error) => {
          console.log(error);
          alert("An error occured");
        });
      }
    },

    deletePack() {
      if (confirm("Are you sure?")) {
        PackService.deleteByID(this.$route.params.id)
        .then((response) => {
          console.log(response.data);
          alert("Done");
          this.$router.push({name: "search"});
        })
        .catch((error) => {
          console.log(error);
          alert("An error occured");
        });
      }
    },
  },

  mounted() {
    PackService.searchByID(this.$route.params.id)
      .then((response) => {
        console.log(response);
        this.packmetadata.Name = response.data.metadata.Name;
        this.packmetadata.Version = response.data.metadata.Version;
        this.packmetadata.ID = response.data.metadata.ID;
        this.packdata.JSProgram = response.data.data.JSProgram;
        this.packdata.Content = response.data.data.Content;
        this.loaded = true;
      })
      .catch((error) => {
        console.log(error);
        alert("An error occured");
      });
  },
});
</script>