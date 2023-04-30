import axios, { AxiosInstance } from 'axios';
import {
  PackageData,
  Package,
  PackageQuery,
  PackageRegEx,
} from "@/types/Pack"
/* eslint-disable */
class PackService {
  private http: AxiosInstance;

  constructor() {
    this.http = axios.create({
      baseURL: "https://pjbackend-7ytde3snta-uc.a.run.app",
      headers: { "Content-type": "application/json" },
    });
  }

  getPack(offset: number, data: PackageQuery[]): Promise<any> {
    return this.http.post(`/packages?offset=${offset}`, data);
  }

  searchByID(id: any): Promise<any> {
    return this.http.get(`/package/${id}`);
  }

  updateByID(id: any, data: Package): Promise<any> {
    return this.http.put(`/package/${id}`, data);
  }

  deleteByID(id: any): Promise<any> {
    return this.http.delete(`/package/${id}`);
  }

  npmIngest(data: PackageData): Promise<any> {
    return this.http.post("/package", data);
  }

  ratePack(id: any): Promise<any> {
    return this.http.get(`/package/${id}/rate`);
  }

  // searchByName(name: string): Promise<any> {
  //   return http.get(`/package/${name}`);
  // }

  deleteByName(name: string): Promise<any> {
    return this.http.delete(`/package/${name}`);
  }

  searchByRegex(data: PackageRegEx): Promise<any> {
    return this.http.post(`/package/byRegEx`, data);
  }

  resetRegistry(): Promise<any> {
    return this.http.delete(`/reset`);
  }
}

export default new PackService();
