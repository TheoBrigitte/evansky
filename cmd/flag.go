package cmd

func init() {
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "show less informations")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more informations")
}
