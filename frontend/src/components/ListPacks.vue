<template>
  <div id="body-wrapper">
    <div class="input-group">
      <input type="text" class="form-control" placeholder="Name" v-model="name"/>
      <input v-if="!useRegex" type="text" class="form-control" placeholder="Version" v-model="version"/>
      <button class="btn btn-ada" type="button" @click="searchPack">Search</button>
    </div>
    <br>
    <input id="regexCheckbox" type="checkbox" label="Regex switch checkbox" v-model="useRegex"/>
    <label style="padding-left: 5px" for="regexCheckbox">Use regex</label>
    <div v-if="totalPages > 0">
      <br>
      <div>
        <h4>Matched {{ totalPacks }} Packages</h4>
        <div class="grid-container list-title">
          <div class="shaded-row grid-item first-column">Name</div>
          <div class="shaded-row grid-item">Version</div>
        </div>
        <ul class="multi-row-list">
          <li>
            <div class="grid-container list-row" v-for="(pack, i) in packs" :key="i">
              <div class="grid-item first-column" :class="{'shaded-row': i % 2}">
                <router-link :to="'/package/' + pack.ID">{{ pack.Name }}</router-link> 
            </div>
              <div class="grid-item" :class="{'shaded-row': i % 2}"> {{ pack.Version }} </div>
            </div>
          </li>
        </ul>
      </div>
      <br>
      <nav aria-label="Page navigation" v-if="totalPages > 1">
        <ul class="page-nav-bar pagination">
          <li class="page-item" :class="{ disabled: currentPage === 1 }">
            <button class="page-link" @click="paginate(currentPage - 1)">Previous</button>
          </li>
          <li class="page-item" :class="{ active: index + 1 === currentPage }" v-for="(_, index) in totalPages + 1" :key="index">
            <button class="page-link" @click="paginate(index + 1)">{{ index + 1 }}</button>
          </li>
          <li class="page-item" :class="{ disabled: currentPage === totalPages }">
            <button class="page-link" @click="paginate(currentPage + 1)">Next</button>
          </li>
        </ul>
      </nav>
    </div>
    <div v-else>
      <br>
      <h4>Matched 0 Packages</h4>
    </div>
  </div>
  <footer class="fixed-bottom">
    <div class="text-center p-3" style="background-color: rgba(0, 0, 0, 0.2);">
      Built with Vue.js, npmjs, node, and Bootstrap.js
    </div>
  </footer>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import PackService from "@/services/PackService";
import { 
  PackageMetadata,
  PackageQuery,
  PackageRegEx,
} from "@/types/Pack";

export default defineComponent({
  name: "list-packages",
  data() {
    return {
      packs: [] as PackageMetadata[],
      name: "",
      version: "",
      currentPage: 1,
      totalPages: 0,
      totalPacks: 0,
      useRegex: false,
    };
  },
  methods: {
    retrievePackages(page: number) {
      const offset = page;

      if (!this.useRegex) {
        const data: PackageQuery[] = [{
          Name: this.name === "" ? "*" : this.name,
          Version: this.version,
        }];

        PackService.getPack(offset, data)
          .then((response) => {
            this.packs = response.data;
            this.totalPages = response.headers['page-count'];
            this.totalPacks = response.headers['pack-count'];
          })
          .catch((error) => {
            alert(error.response.data.message);
            window.location.reload();
        });
      } else {
        const data: PackageRegEx = {
          RegEx: this.name,
        };

        PackService.searchByRegex(data)
          .then((response) => {
            this.packs = response.data;
            this.totalPages = response.headers['page-count'];
            this.totalPacks = response.headers['pack-count'];
          })
          .catch((error) => {
            alert(error.response.data.message);
            window.location.reload();
        });
      }
    },

    searchPack() {
       this.retrievePackages(1);
       this.currentPage = 1;
     },

    paginate(page: number) {
      if (page < 1 || page > this.totalPages) {
        return;
      }

      this.currentPage = page;
      this.retrievePackages(this.currentPage);
    },
  },
  
  mounted() {
    this.totalPages = 0;
    this.totalPacks = 0;
    this.useRegex = false;
    this.retrievePackages(this.currentPage);
  },
});
</script>

<style>
@import '@/assets/style.css';
</style>