<template>
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
          v-model="pack.JSProgram"
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
          v-model="pack.URL"
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
</template>

<script lang="ts">
import { defineComponent } from "vue";
import PackService from "@/services/PackService";
import { 
  PackageData,
  Package,
} from "@/types/Pack";

export default defineComponent({
  name: "add-pack",
  data() {
    return {
      pack: {} as PackageData,
      resourceType: "url",
      submitted: false,
    };
  },
  methods: {
    async savePack() {
      let data: any = {
        JSProgram: this.pack.JSProgram,
      };

      if (this.resourceType === "url") {
        data.URL = this.pack.URL;
      } else {
        const zipInput = this.$refs.zipInput as HTMLInputElement;
        const file = zipInput.files?.[0];

        if (file) {
          const fileReader = new FileReader();
          await new Promise((resolve, reject) => {
            fileReader.onload = () => {
              // Get the base64 string of the uploaded zip file
              data.Content = fileReader.result as string;
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
      PackService.npmIngest(data)
        .then((response) => {
          console.log(response.data.metadata.Name);
          console.log(response.data.metadata.Version);
          console.log(response.data.metadata.ID);
          alert("Done, redirecting");
          this.$router.push({name: "search"});
        })
        .catch((error) => {
          alert("An error occured");
          console.log(error);
          window.location.reload();
        });
    },

    newPack() {
      this.submitted = false;
      this.pack = {} as PackageData;
    },
  },
});
</script>

<style>
@import '@/assets/style.css';
</style>
